# Core CRDT Specs

The following is a specification for the currently supported CRDTs in DefraDB.

## Overview
CRDTs are Conflict-Free Replicated Data Types, which is a type of data structure that automatically handles conflicts from multiple writers in a deterministic manner. There are many types of CRDTs, each with their own semantics, use case, pros, and cons.

Additionally, there are many different classes of CRDTs. DefraDB uses delta-state CRDTs, and is intended to be used with MerkleClocks to create a Merkle CRDT. Check [this](https://arxiv.org/abs/1603.01529) for an overview of delta-state CRDTs. Check [this](https://hector.link/presentations/merkle-crdts/merkle-crdts.pdf) for an overview of Merkle CRDTs.

## State
Each CRDT implemented in this /core/crdt  package is a `stateless` CRDT. Meaning, all of the state is internal to the provided Datastore object, almost nothing resides in memory in the ```struct```.

All CRDTs need some indication of when some event (update) happened relative to other events. In Distributed Systems this is called a `Clock`. There are various types, including Vector, Logical, or Wall Clocks. The notable difference of Merkle CRDTs, is they use a Merkle Clock to track events.Merkle Clocks keep track of a ```priority``` value each event occurred at.

In the Datastore, everything is saved as Key-Value pairs. So if we had a CRDT identified by ID: `mydata`, then the following K/V structure would exist. Below is an example of a ```LWWRegister``` K/V layout:
```
/mydata:v => Value
/mydata:p => Priority
```
Where **Value** is the current serialized and joined value of the CRDT, and **Priority** is the current priority value the last update occurred at. 

CRDT Key-Value state can always be namespaced by some arbitrary value with conflicting with the data type semantics. Eg. ```/somenamespace/mydata:v``` would work the exact same in the above example.

Currently, all on-disk **Value**'s are serialized byte arrays using the [CBOR]() encoding. **Priority**'s are binary serialized uint64.

## Deltas
Delta-State CRDT are a unique class of CRDTs, that are very efficient for Merkle CRDTs. Delta-State CRDTs operate on a collection of deltas, a delta being a point in time representation of an intended action against a CRDT (Ie: increment counter, set register, remove element, etc.). Additionally, the CRDT Mutator functions, where traditionally operate on state, instead are Delta Mutators, which return a delta for the given state. 

## CRDTs

Here is a breakdown of all the currently supported CRDTs, and their underlying semantics.

### LWWRegister - Last-Write-Wins Register
Registers are the most basic form of CRDT, as they simple represent any arbitrary value, like and integer, or string. A Last-Write-Win semantic means that when merging two Registers, in the event of a conflict, the update with the latest write time according the given clock, is the chosen value (winner). For Merkle Clocks, this is the event with the highest ```priority``` value.

#### Methods
```
- Set(value []byte) -> Delta # Return a new Delta object with the given value

- Value() -> ([]byte, error) # Returns the current serialized state

- Merge(delta) -> error # Merges two LWWRegister deltas together.
```

#### Semantics
Any update to a Last Write Win Register always creates a conflict, since its only a single value. To resolve the conflict, the delta with the highest ```priority``` value is chosen. If two deltas have the same ```priority``` then the highest lexicographic value of the delta wins.

#### Key-Value Layout
Since Registers are simplistic by design, their k/v layout is also simple.
With a Register identified by ```myregister```:
```
/myregister:v => Value
/myregister:p => Priorty
```

### GCounter - Increment-Only Counter
Counters allow for an integer (or float) to be updated over time via basic ```increment``` methods. They can be used for a number of scenarios, like view counter, user followers, etc. An Increment-Only counter means you can only ever increase the stored value, not decrease, see **PNCounter** to include decrement operations.


#### Methods
```
- increment(val uint64) -> Delta # Return a new Delta object with the increment operation

- Value() -> (uint64, error) # Returns the current counter value

- Merge(delta) -> error # Merge the current state with a new delta
```

#### Semantics
New increment operations are resolved using a ```Max()``` function on conflicting deltas (those with equal priority values). *Common implementations of the GCounter use a map of Counter ReplicaIds to store multiple counters in a single GCounter, Defra does not use this method, instead only storing a single counter in a single GCounter*.

#### Key-Value Layout
With a GCounter identified by ```mygcounter```
```
/mygcounter:v => Value
/mygcounter:p => Priority
```

### PNCounter - Increment/Decrement Counter
A PNCounter is equivalent to the GCounter, with the notable exception it can be incremented and decremented. This is achieved by composing a PNCounter from two individual GCounters, one to track increment ops, the other to track decrement ops. `PNCounter := GCounter + GCounter.`

#### Methods
```
- Increment(val uint64) -> Delta # Return a new Delta with the increment operation value

- Decrement(val uint64) -> Delta # Return a new Delta with the decrement operation value

- Value() -> (int64, error) -> # Returns the current counter value which is the summation of the increment counter set and the decrement counter set.

- Merge(delta) -> # Merge the current state with a new delta
```

#### Semantics
Internally, the PNCounter uses two GCounters, so all the same semantics are carried over. See [GCounter Semantics](#GCounter)

#### Key-Value Layout
With a PNCounter identified by ```mypncounter```
```
/mypncounter/inc:v => Value
/mypncounter/dec:v => Value
/mypncounter:p => Priority
```
You'll notice that despite the PNCounter constructed from two GCounter, they use a shared ```Priority``` Key-Value pair, this isn't required, but it reduces redundancy. It also shares a single ```Head``` value for its associated MerkleClock, which is covered in the [Merkle CRDT]() section.

### EW-Flag - Enable-Wins Flag

### DW-Flag -  Disable Wins Flag

### LWWW-Set - Last-Write-Wins Set

### OR-Set - Add-Wins Observe-Remove Set

### LWW-Map - Last-Write-Wins Map

### OR-Map - Add-Wins Observe-Remove Map
An AWORMap is a Map like CRDT structure, in that we store keys and values. The values are themselves CRDTs of any kind (more on this later), and in the face of a conflict, the addition of a key wins. Keys are basic string identifiers, and values are CRDTs, so any further conflict can be handled by their respective CRDT semantics of the Value.



