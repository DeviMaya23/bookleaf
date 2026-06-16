const SIZE = 64
const DCT_SIZE = 8

// Precomputed cosine table for separable DCT: cosTable[n][k] = cos(π·k·(2n+1)/(2·SIZE))
// Computed once at module load; avoids repeated Math.cos calls in the hot loop.
const cosTable: number[][] = Array.from({ length: SIZE }, (_, n) =>
  Array.from({ length: DCT_SIZE }, (_, k) =>
    Math.cos((Math.PI * k * (2 * n + 1)) / (2 * SIZE)),
  ),
)

export function computePHash(canvas: OffscreenCanvas): string {
  // Draw to 64×64. Read raw RGB channels to apply Rec. 601 luminance
  // (matching goimagehash: 0.299R + 0.587G + 0.114B).
  const smallCanvas = new OffscreenCanvas(SIZE, SIZE)
  const ctx = smallCanvas.getContext('2d')!
  ctx.drawImage(canvas, 0, 0, SIZE, SIZE)

  const { data } = ctx.getImageData(0, 0, SIZE, SIZE)
  const pixels = new Float64Array(SIZE * SIZE)
  for (let i = 0; i < SIZE * SIZE; i++) {
    const o = i * 4
    pixels[i] = 0.299 * data[o] + 0.587 * data[o + 1] + 0.114 * data[o + 2]
  }

  // Separable 2D DCT — compute only the top-left 8×8 of the 64×64 DCT output.
  // Step 1: 1D DCT over each row, keep first DCT_SIZE coefficients.
  const rowDct = new Float64Array(SIZE * DCT_SIZE)
  for (let row = 0; row < SIZE; row++) {
    const base = row * SIZE
    const out = row * DCT_SIZE
    for (let k = 0; k < DCT_SIZE; k++) {
      let sum = 0
      for (let n = 0; n < SIZE; n++) {
        sum += pixels[base + n] * cosTable[n][k]
      }
      rowDct[out + k] = sum
    }
  }

  // Step 2: 1D DCT over each of the first 8 columns of the row-transformed
  // matrix, keep first DCT_SIZE coefficients → 8×8 = 64 final values.
  const vals = new Float64Array(DCT_SIZE * DCT_SIZE)
  for (let col = 0; col < DCT_SIZE; col++) {
    for (let k = 0; k < DCT_SIZE; k++) {
      let sum = 0
      for (let n = 0; n < SIZE; n++) {
        sum += rowDct[n * DCT_SIZE + col] * cosTable[n][k]
      }
      vals[k * DCT_SIZE + col] = sum
    }
  }

  // Median of all 64 values (including DC at [0][0]), matching goimagehash.
  const sorted = Array.from(vals).sort((a, b) => a - b)
  const median = (sorted[31] + sorted[32]) / 2

  return Array.from(vals).map(v => (v > median ? '1' : '0')).join('')
}
