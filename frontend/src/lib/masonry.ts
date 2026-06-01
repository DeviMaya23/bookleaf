export const MASONRY_TARGET_COL_WIDTH = 220
export const GAP = 12

export function computeMasonryLayout(containerWidth: number) {
  const numCols = Math.max(1, Math.floor(containerWidth / MASONRY_TARGET_COL_WIDTH))
  const colWidth = (containerWidth - GAP * (numCols - 1)) / numCols
  return { numCols, colWidth }
}
