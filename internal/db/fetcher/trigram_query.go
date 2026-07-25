// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package fetcher

import (
	"cmp"
	"regexp/syntax"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

// This file turns a filter condition on a trigram-indexed field into a boolean query over
// trigrams, which the trigram index iterator evaluates against the stored postings.
//
// The regular expression half is derived from index/regexp.go of
// https://github.com/google/codesearch (BSD-3), described by Russ Cox in
// https://swtch.com/~rsc/regexp/regexp4.html. Two differences from the original:
//
//   - Literals are lowercased as they are read out of the parsed expression, because the
//     index stores lowercased trigrams. A case-sensitive filter therefore over-returns and
//     is narrowed again by the recheck.
//   - The boolean query is only ever a filter, so the query printer and the parts of the
//     original that serve its own on-disk format are left out.
//
// The result is always conservative: it matches everything the condition matches and
// usually a good deal more. Every candidate it produces is rechecked against the original
// filter by filteredFetcher, so over-returning costs time and never correctness.

// trigramQuery is a boolean query over trigrams.
type trigramQuery struct {
	op      trigramQueryOp
	trigram []string
	sub     []*trigramQuery
}

type trigramQueryOp int

const (
	// trigramAll matches every document: the condition yielded nothing the index can use,
	// so the caller should scan instead.
	trigramAll trigramQueryOp = iota
	// trigramNone matches no document at all.
	trigramNone
	// trigramAnd requires every trigram and sub-query to match.
	trigramAnd
	// trigramOr requires at least one trigram or sub-query to match.
	trigramOr
)

var allQuery = &trigramQuery{op: trigramAll}
var noneQuery = &trigramQuery{op: trigramNone}

// likeTrigramQuery returns the trigram query for a _like or _ilike pattern.
//
// Both operators are served by the same query: the index holds lowercased trigrams, so the
// case-sensitive form simply over-returns and is narrowed by the recheck.
func likeTrigramQuery(pattern string) *trigramQuery {
	q := allQuery
	for _, literal := range likeLiterals(pattern) {
		q = q.andTrigrams(stringSet{strings.ToLower(literal)})
	}
	return q
}

// likeLiterals returns the literal runs that a value must contain to satisfy the given _like
// pattern. It mirrors the shapes that connor's like operator recognises, so a pattern it
// treats as a plain string (one with '%' in more than one interior position, for example)
// is required here in full, '%' included.
func likeLiterals(pattern string) []string {
	hasPrefix := false
	hasSuffix := false
	var startAndEnd []string

	if len(pattern) >= 2 {
		if pattern[0] == '%' {
			hasPrefix = true
			pattern = strings.TrimPrefix(pattern, "%")
		}
		if pattern[len(pattern)-1] == '%' {
			hasSuffix = true
			pattern = strings.TrimSuffix(pattern, "%")
		}
		if !hasPrefix && !hasSuffix {
			startAndEnd = strings.Split(pattern, "%")
		}
	}

	// '%x%' requires x anywhere, '%x' requires it at the end and 'x%' at the start. All three
	// require the value to contain x, which is all a trigram index can express.
	if !hasPrefix && !hasSuffix && len(startAndEnd) == 2 {
		return startAndEnd
	}
	return []string{pattern}
}

// regexTrigramQuery returns the trigram query for a _regex pattern.
//
// A pattern that does not parse yields trigramAll, so the query falls back to a scan and the
// filter reports the invalid pattern itself, exactly as it does with no index present.
func regexTrigramQuery(pattern string) *trigramQuery {
	re, err := syntax.Parse(pattern, syntax.Perl)
	if err != nil {
		return allQuery
	}
	info := analyzeRegexp(re.Simplify())
	info.simplify(true)
	info.addExact()
	return info.match
}

// and returns the query q AND r, possibly reusing q's and r's storage.
func (q *trigramQuery) and(r *trigramQuery) *trigramQuery {
	return q.andOr(r, trigramAnd)
}

// or returns the query q OR r, possibly reusing q's and r's storage.
func (q *trigramQuery) or(r *trigramQuery) *trigramQuery {
	return q.andOr(r, trigramOr)
}

// andOr returns the query q AND r or q OR r, possibly reusing q's and r's storage.
// It works hard to avoid creating unnecessarily complicated structures.
func (q *trigramQuery) andOr(r *trigramQuery, op trigramQueryOp) *trigramQuery {
	if len(q.trigram) == 0 && len(q.sub) == 1 {
		q = q.sub[0]
	}
	if len(r.trigram) == 0 && len(r.sub) == 1 {
		r = r.sub[0]
	}

	// Boolean simplification.
	// If q implies r, q AND r is q, and q OR r is r.
	if q.implies(r) {
		if op == trigramAnd {
			return q
		}
		return r
	}
	if r.implies(q) {
		if op == trigramAnd {
			return r
		}
		return q
	}

	// Both q and r are an AND or an OR. If they match, or can be made to match, merge them.
	qAtom := len(q.trigram) == 1 && len(q.sub) == 0
	rAtom := len(r.trigram) == 1 && len(r.sub) == 0
	if q.op == op && (r.op == op || rAtom) {
		q.trigram = stringSet.union(q.trigram, r.trigram, false)
		q.sub = append(q.sub, r.sub...)
		return q
	}
	if r.op == op && qAtom {
		r.trigram = stringSet.union(r.trigram, q.trigram, false)
		return r
	}
	if qAtom && rAtom {
		q.op = op
		q.trigram = append(q.trigram, r.trigram...)
		return q
	}

	// If one matches the op, add the other to it.
	if q.op == op {
		q.sub = append(q.sub, r)
		return q
	}
	if r.op == op {
		r.sub = append(r.sub, q)
		return r
	}

	// We are creating an AND of ORs or an OR of ANDs.
	return &trigramQuery{op: op, sub: []*trigramQuery{q, r}}
}

// implies reports whether q implies r. It is allowed to return false negatives, and only
// answers the two cases that keep trigramAll and trigramNone from being nested inside a
// query: a nested trigramAll node constrains nothing, and isEvaluable rejects it.
func (q *trigramQuery) implies(r *trigramQuery) bool {
	if q.op == trigramNone || r.op == trigramAll {
		// False implies everything, and everything implies true.
		return true
	}
	// True implies nothing, and nothing implies false. Anything else is reported as no
	// implication, which costs a larger query and never a wrong one.
	return false
}

// andTrigrams returns q AND the OR of the AND of the trigrams present in each string of t.
func (q *trigramQuery) andTrigrams(t stringSet) *trigramQuery {
	if t.minLen() < 3 {
		// The query has to hold for every string in the set, so one member too short to have
		// a trigram means no trigram is required at all. This is also how the empty string in
		// a prefix or suffix set propagates "nothing is known" without any extra bookkeeping,
		// and removing this guard is what would turn the result into false negatives.
		return q
	}

	or := noneQuery
	for _, tt := range t {
		var trig stringSet
		for i := 0; i+3 <= len(tt); i++ {
			trig.add(tt[i : i+3])
		}
		trig.clean(false)
		or = or.or(&trigramQuery{op: trigramAnd, trigram: trig})
	}
	return q.and(or)
}

// String renders the query in the notation used by the reference implementation:
// juxtaposition is AND, '|' is OR, '+' matches everything and '-' matches nothing.
// It exists for the tests, which are far easier to read and to compare against the
// reference in this form than as a tree of structs.
func (q *trigramQuery) String() string {
	if q == nil {
		return "?"
	}
	if q.op == trigramNone {
		return "-"
	}
	if q.op == trigramAll {
		return "+"
	}
	if len(q.sub) == 0 && len(q.trigram) == 1 {
		return strconv.Quote(q.trigram[0])
	}

	var s, sjoin, end, tjoin string
	if q.op == trigramAnd {
		sjoin = " "
		tjoin = " "
	} else {
		s = "("
		sjoin = ")|("
		end = ")"
		tjoin = "|"
	}
	for i, t := range q.trigram {
		if i > 0 {
			s += tjoin
		}
		s += strconv.Quote(t)
	}
	if len(q.sub) > 0 {
		if len(q.trigram) > 0 {
			s += sjoin
		}
		s += q.sub[0].String()
		for i := 1; i < len(q.sub); i++ {
			s += sjoin + q.sub[i].String()
		}
	}
	return s + end
}

// regexpInfo summarizes the results of analyzing a regular expression.
type regexpInfo struct {
	// canEmpty records whether the expression matches the empty string.
	canEmpty bool

	// exact is the exact set of strings matching the expression, or nil if it is not known.
	exact stringSet

	// if exact is nil, prefix is the set of possible match prefixes and suffix is the set of
	// possible match suffixes.
	prefix stringSet
	suffix stringSet

	// match is a query that any match for the expression must satisfy, in addition to the
	// information recorded above.
	match *trigramQuery
}

const (
	// Exact sets are limited to maxExact strings. If they get bigger, simplify rewrites the
	// regexpInfo to use prefix and suffix instead.
	maxExact = 7

	// Prefix and suffix sets are limited to maxSet strings. If they get bigger, simplify
	// replaces groups of strings sharing a common leading prefix (or trailing suffix) with
	// that prefix (or suffix).
	//
	// Both bounds are a trust boundary rather than a tuning knob: the expression arrives in a
	// query, and the set operations below are cross products, so without them a pattern like
	// (a|b|c)(d|e|f)(g|h|i)... costs the server an unbounded amount of work.
	maxSet = 20
)

// anyMatch returns the regexpInfo of an expression that matches any string.
func anyMatch() regexpInfo {
	return regexpInfo{
		canEmpty: true,
		prefix:   []string{""},
		suffix:   []string{""},
		match:    allQuery,
	}
}

// anyChar returns the regexpInfo of an expression that matches any single character.
func anyChar() regexpInfo {
	return regexpInfo{
		prefix: []string{""},
		suffix: []string{""},
		match:  allQuery,
	}
}

// noMatch returns the regexpInfo of an expression that matches no string at all.
func noMatch() regexpInfo {
	return regexpInfo{match: noneQuery}
}

// emptyString returns the regexpInfo of an expression that matches only the empty string.
func emptyString() regexpInfo {
	return regexpInfo{
		canEmpty: true,
		exact:    []string{""},
		match:    allQuery,
	}
}

// analyzeRegexp returns the regexpInfo for the parsed expression re.
//
// Literal text is lowercased on the way in, so every set below holds strings in the form the
// index stores them and no later step has to remember to fold again.
func analyzeRegexp(re *syntax.Regexp) regexpInfo {
	var info regexpInfo
	switch re.Op {
	case syntax.OpNoMatch:
		return noMatch()

	case syntax.OpEmptyMatch,
		syntax.OpBeginLine, syntax.OpEndLine,
		syntax.OpBeginText, syntax.OpEndText,
		syntax.OpWordBoundary, syntax.OpNoWordBoundary:
		// An anchor matches the empty string, so anchoring adds nothing to the query. It costs
		// nothing either: '^abc' still requires the trigram "abc".
		return emptyString()

	case syntax.OpLiteral:
		if re.Flags&syntax.FoldCase != 0 {
			switch len(re.Rune) {
			case 0:
				return emptyString()
			case 1:
				// Single case-folded letter: rewrite as the character class of its folding
				// orbit and analyze that. Lowercasing alone would not do here, since a rune
				// can fold to a letter it does not lowercase to.
				re1 := &syntax.Regexp{Op: syntax.OpCharClass}
				re1.Rune = re1.Rune0[:0]
				r0 := re.Rune[0]
				re1.Rune = append(re1.Rune, r0, r0)
				for r1 := unicode.SimpleFold(r0); r1 != r0; r1 = unicode.SimpleFold(r1) {
					re1.Rune = append(re1.Rune, r1, r1)
				}
				return analyzeRegexp(re1)
			}
			// Several case-folded letters: treat as the concatenation of single ones.
			re1 := &syntax.Regexp{Op: syntax.OpLiteral, Flags: syntax.FoldCase}
			info = emptyString()
			for i := range re.Rune {
				re1.Rune = re.Rune[i : i+1]
				info = concatRegexpInfo(info, analyzeRegexp(re1))
			}
			return info
		}
		info.exact = stringSet{strings.ToLower(string(re.Rune))}
		info.match = allQuery

	case syntax.OpAnyCharNotNL, syntax.OpAnyChar:
		return anyChar()

	case syntax.OpCapture:
		return analyzeRegexp(re.Sub[0])

	case syntax.OpConcat:
		return foldRegexpInfo(concatRegexpInfo, re.Sub, emptyString())

	case syntax.OpAlternate:
		return foldRegexpInfo(alternateRegexpInfo, re.Sub, noMatch())

	case syntax.OpQuest:
		return alternateRegexpInfo(analyzeRegexp(re.Sub[0]), emptyString())

	case syntax.OpStar:
		// x* can match nothing at all, so everything known about x is discarded.
		return anyMatch()

	case syntax.OpRepeat:
		if re.Min == 0 {
			// Like OpStar.
			return anyMatch()
		}
		fallthrough
	case syntax.OpPlus:
		// There has to be at least one x, so the prefixes and suffixes stay the same. If x
		// was exact, it is not anymore.
		info = analyzeRegexp(re.Sub[0])
		if info.exact.have() {
			info.prefix = info.exact
			info.suffix = slices.Clone(info.exact)
			info.exact = nil
		}

	case syntax.OpCharClass:
		info.match = allQuery

		if len(re.Rune) == 0 {
			return noMatch()
		}

		if len(re.Rune) == 1 {
			info.exact = stringSet{strings.ToLower(string(re.Rune[0]))}
			break
		}

		n := 0
		for i := 0; i < len(re.Rune); i += 2 {
			n += int(re.Rune[i+1] - re.Rune[i])
		}
		// A large class could be enumerated, but overestimating it is allowed and much
		// cheaper. This is what makes '.' and a negated class give up.
		if n > 100 {
			return anyChar()
		}

		info.exact = []string{}
		for i := 0; i < len(re.Rune); i += 2 {
			lo, hi := re.Rune[i], re.Rune[i+1]
			for rr := lo; rr <= hi; rr++ {
				info.exact.add(strings.ToLower(string(rr)))
			}
		}
	}

	info.simplify(false)
	return info
}

// foldRegexpInfo is the usual higher-order function.
func foldRegexpInfo(
	f func(x, y regexpInfo) regexpInfo,
	sub []*syntax.Regexp,
	zero regexpInfo,
) regexpInfo {
	if len(sub) == 0 {
		return zero
	}
	if len(sub) == 1 {
		return analyzeRegexp(sub[0])
	}
	info := f(analyzeRegexp(sub[0]), analyzeRegexp(sub[1]))
	for i := 2; i < len(sub); i++ {
		info = f(info, analyzeRegexp(sub[i]))
	}
	return info
}

// concatRegexpInfo returns the regexpInfo for xy given x and y.
func concatRegexpInfo(x, y regexpInfo) regexpInfo {
	var xy regexpInfo
	xy.match = x.match.and(y.match)
	if x.exact.have() && y.exact.have() {
		xy.exact = x.exact.cross(y.exact, false)
	} else {
		if x.exact.have() {
			xy.prefix = x.exact.cross(y.prefix, false)
		} else {
			xy.prefix = x.prefix
			if x.canEmpty {
				xy.prefix = xy.prefix.union(y.prefix, false)
			}
		}
		if y.exact.have() {
			xy.suffix = x.suffix.cross(y.exact, true)
		} else {
			xy.suffix = y.suffix
			if y.canEmpty {
				xy.suffix = xy.suffix.union(x.suffix, true)
			}
		}
	}

	// If every string in the cross product of x's suffixes and y's prefixes is long enough,
	// then one of their trigrams must be present, and nothing else would have recorded it.
	if !x.exact.have() && !y.exact.have() &&
		len(x.suffix) <= maxSet && len(y.prefix) <= maxSet &&
		x.suffix.minLen()+y.prefix.minLen() >= 3 {
		xy.match = xy.match.andTrigrams(x.suffix.cross(y.prefix, false))
	}

	xy.simplify(false)
	return xy
}

// alternateRegexpInfo returns the regexpInfo for x|y given x and y.
func alternateRegexpInfo(x, y regexpInfo) regexpInfo {
	var xy regexpInfo
	switch {
	case x.exact.have() && y.exact.have():
		xy.exact = x.exact.union(y.exact, false)
	case x.exact.have():
		xy.prefix = x.exact.union(y.prefix, false)
		xy.suffix = x.exact.union(y.suffix, true)
		x.addExact()
	case y.exact.have():
		xy.prefix = x.prefix.union(y.exact, false)
		xy.suffix = x.suffix.union(slices.Clone(y.exact), true)
		y.addExact()
	default:
		xy.prefix = x.prefix.union(y.prefix, false)
		xy.suffix = x.suffix.union(y.suffix, true)
	}
	xy.canEmpty = x.canEmpty || y.canEmpty
	xy.match = x.match.or(y.match)

	xy.simplify(false)
	return xy
}

// addExact adds the trigrams of info.exact to the match query.
func (info *regexpInfo) addExact() {
	if info.exact.have() {
		info.match = info.match.andTrigrams(info.exact)
	}
}

// simplify simplifies the regexpInfo when the exact set gets too large.
func (info *regexpInfo) simplify(force bool) {
	// If there are now too many exact strings, harvest their trigrams and move what is left
	// of them into the prefix and suffix sets. Information is moved, never dropped.
	info.exact.clean(false)
	if len(info.exact) > maxExact || (info.exact.minLen() >= 3 && force) || info.exact.minLen() >= 4 {
		info.addExact()
		for _, s := range info.exact {
			n := len(s)
			if n < 3 {
				info.prefix.add(s)
				info.suffix.add(s)
			} else {
				info.prefix.add(s[:2])
				info.suffix.add(s[n-2:])
			}
		}
		info.exact = nil
	}

	if !info.exact.have() {
		info.simplifySet(&info.prefix)
		info.simplifySet(&info.suffix)
	}
}

// simplifySet reduces the size of the given prefix or suffix set. There is no point carrying
// an enormous set around, since it is only ever used to make trigrams, so as it grows
// simplifySet moves what it holds into the match query instead.
func (info *regexpInfo) simplifySet(s *stringSet) {
	t := *s
	t.clean(s == &info.suffix)

	// Harvest the trigrams of the current set before shortening it.
	info.match = info.match.andTrigrams(t)

	for n := 3; n == 3 || len(t) > maxSet; n-- {
		// Replace the set by strings of length n-1.
		w := 0
		for _, str := range t {
			if len(str) >= n {
				if s == &info.prefix {
					str = str[:n-1]
				} else {
					str = str[len(str)-n+1:]
				}
			}
			if w == 0 || t[w-1] != str {
				t[w] = str
				w++
			}
		}
		t = t[:w]
		t.clean(s == &info.suffix)
	}

	// Drop redundant entries: if "ab" is a possible prefix then "abc" adds nothing.
	w := 0
	f := strings.HasPrefix
	if s == &info.suffix {
		f = strings.HasSuffix
	}
	for _, str := range t {
		if w == 0 || !f(str, t[w-1]) {
			t[w] = str
			w++
		}
	}
	*s = t[:w]
}

// stringSet is a set of strings. The nil stringSet means there is no set (nothing is known);
// the non-nil but empty stringSet is the empty set.
type stringSet []string

// have reports whether there is a set at all.
func (s stringSet) have() bool {
	return s != nil
}

// compareBySuffix orders strings by their last character first, so that the redundancy check
// in simplifySet sees strings sharing a suffix as one sorted run.
func compareBySuffix(s, t string) int {
	for i := 1; i <= len(s) && i <= len(t); i++ {
		if c := cmp.Compare(s[len(s)-i], t[len(t)-i]); c != 0 {
			return c
		}
	}
	return cmp.Compare(len(s), len(t))
}

// add adds str to the set.
func (s *stringSet) add(str string) {
	*s = append(*s, str)
}

// clean sorts the set and removes duplicates.
func (s *stringSet) clean(isSuffix bool) {
	if isSuffix {
		slices.SortFunc(*s, compareBySuffix)
	} else {
		slices.Sort(*s)
	}
	*s = slices.Compact(*s)
}

// minLen returns the length of the shortest string in s.
func (s stringSet) minLen() int {
	if len(s) == 0 {
		return 0
	}
	m := len(s[0])
	for _, str := range s {
		if m > len(str) {
			m = len(str)
		}
	}
	return m
}

// union returns the union of s and t, reusing s's storage.
func (s stringSet) union(t stringSet, isSuffix bool) stringSet {
	s = append(s, t...)
	s.clean(isSuffix)
	return s
}

// cross returns the cross product of s and t, each pair concatenated.
func (s stringSet) cross(t stringSet, isSuffix bool) stringSet {
	p := stringSet{}
	for _, ss := range s {
		for _, tt := range t {
			p.add(ss + tt)
		}
	}
	p.clean(isSuffix)
	return p
}
