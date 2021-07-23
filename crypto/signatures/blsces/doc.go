/*
The BLSCES Package implements an Anonymous Credential system
called BLS CES, which builds a Content Extraction Signature
using the BLS Aggregate Signature system.

Related Papers
BLS: https://www.iacr.org/archive/asiacrypt2001/22480516.pdf
Aggregate BLS: https://crypto.stanford.edu/~dabo/pubs/papers/aggreg.pdf
CES: https://cpb-us-w2.wpmucdn.com/sites.uab.edu/dist/a/68/files/2020/01/cesproc-icisc01-p285.pdf
BLSCES: https://arxiv.org/pdf/2006.05201.pdf

The BLSCES allows us to create anonymous credentials from some
set of data (messages), signed and issued by some actor's
public key to a Holder. The Holder can then Extract a subset
of those messages to present to any requesting party, with
cryptographic proof that the subset of presented data is
from a signed and correctly issued whole.
*/
package blsces
