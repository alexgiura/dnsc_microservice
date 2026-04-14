/** Gap între trigger și panou; margin față de marginile viewport-ului. */
const GAP = 4
const MARGIN = 8

/**
 * Poziționare viewport-aware (flip): preferă jos; dacă nu încape, deschide sus.
 * `maxHeight` permite scroll când lista depășește spațiul disponibil.
 */
export function computeFlipPopoverY(
  triggerRect: DOMRect,
  contentHeight: number
): { top: number; maxHeight: number; placement: 'above' | 'below' } {
  const vh =
    typeof window !== 'undefined'
      ? window.visualViewport?.height ?? window.innerHeight
      : 800

  const spaceBelow = vh - triggerRect.bottom - GAP - MARGIN
  const spaceAbove = triggerRect.top - GAP - MARGIN

  let top: number
  let maxH: number
  let placement: 'above' | 'below'

  if (contentHeight <= spaceBelow) {
    top = triggerRect.bottom + GAP
    maxH = spaceBelow
    placement = 'below'
  } else if (contentHeight <= spaceAbove) {
    top = triggerRect.top - GAP - contentHeight
    maxH = spaceAbove
    placement = 'above'
  } else {
    // Nu încape integral: partea cu mai mult spațiu + scroll în panou
    if (spaceBelow >= spaceAbove) {
      top = triggerRect.bottom + GAP
      maxH = spaceBelow
      placement = 'below'
    } else {
      const visibleH = Math.min(contentHeight, spaceAbove)
      top = triggerRect.top - GAP - visibleH
      maxH = spaceAbove
      placement = 'above'
    }
  }

  const minVisible = 48
  top = Math.max(MARGIN, Math.min(top, vh - MARGIN - minVisible))

  return {
    top,
    maxHeight: Math.max(80, maxH),
    placement,
  }
}
