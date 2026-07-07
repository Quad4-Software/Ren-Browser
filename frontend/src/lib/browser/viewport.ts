export const MOBILE_LAYOUT_MAX_WIDTH = 768;

export function isCompactViewport(width = window.innerWidth): boolean {
  return width <= MOBILE_LAYOUT_MAX_WIDTH;
}
