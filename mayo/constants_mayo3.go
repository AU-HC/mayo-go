//go:build mayo3

package mayo

const N = 118
const M = 108
const sigBytes = 681
const saltBytes = 32
const digestBytes = 48
const pkSeedBytes = 16
const o = 10
const v = N - o
const K = 11
const q = 16
const OBytes = 540
const vBytes = 54
const rBytes = 55
const P1Bytes = M * ((N - o) * ((N - o) + 1) / 2) / 2
const P2Bytes = M * (N - o) * o / 2
const P3Bytes = M * ((o + 1) * o / 2) / 2
const cskBytes = 32
const cpkBytes = pkSeedBytes + P3Bytes
const eskBytes = skSeedBytes + OBytes + P1Bytes + lBytes
const epkBytes = P1Bytes + P2Bytes + P3Bytes
const lBytes = M * (N - o) * o / 2
const skSeedBytes = saltBytes
const tailFLength = 4

const aCols = K*o + 1
const mVecLimbs = 7
const P1Limbs = v * (v + 1) / 2 * mVecLimbs
const P2Limbs = v * o * mVecLimbs
const P3Limbs = o * (o + 1) / 2 * mVecLimbs

var tailF = [4]byte{8, 0, 1, 7}
