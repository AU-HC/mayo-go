//go:build mayo5

package mayo

const N = 154
const M = 142
const sigBytes = 964
const saltBytes = 40
const digestBytes = 64
const pkSeedBytes = 16
const o = 12
const v = N - o
const K = 12
const q = 16
const OBytes = 852
const vBytes = 71
const rBytes = 72
const P1Bytes = M * ((N - o) * ((N - o) + 1) / 2) / 2
const P2Bytes = M * (N - o) * o / 2
const P3Bytes = M * ((o + 1) * o / 2) / 2
const cskBytes = 32
const cpkBytes = pkSeedBytes + P3Bytes
const eskBytes = skSeedBytes + OBytes + P1Bytes + lBytes
const epkBytes = P1Bytes + P2Bytes + P3Bytes
const lBytes = M * (N - o) * o / 2
const skSeedBytes = saltBytes

const aCols = K*o + 1
const mVecLimbs = 9
const P1Limbs = v * (v + 1) / 2 * mVecLimbs
const P2Limbs = v * o * mVecLimbs
const P3Limbs = o * (o + 1) / 2 * mVecLimbs

var tailF = [4]byte{4, 0, 8, 1}
