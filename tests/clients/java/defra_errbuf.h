// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

#ifndef DEFRA_ERRBUF_H
#define DEFRA_ERRBUF_H

#include <stdio.h>

// defra_set_err writes msg into errbuf, truncating (via snprintf) to fit
// within errbufLen and always leaving it NUL-terminated. A NULL errbuf or
// non-positive errbufLen is a no-op, so callers that don't want the error
// detail can pass NULL/0 without needing their own guard.
static inline void defra_set_err(char* errbuf, int errbufLen, const char* msg) {
    if (errbuf == NULL || errbufLen <= 0) {
        return;
    }
    snprintf(errbuf, (size_t)errbufLen, "%s", msg);
}

#endif // DEFRA_ERRBUF_H
