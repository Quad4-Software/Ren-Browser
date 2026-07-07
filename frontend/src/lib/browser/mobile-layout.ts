export type MobileLayoutOverride = "mobile" | "desktop" | null;

export function resolveMobileUI(input: {
  layoutOverride: MobileLayoutOverride;
  isMobilePlatform: boolean;
  compactViewport: boolean;
}): boolean {
  if (input.layoutOverride === "mobile") {
    return true;
  }
  if (input.layoutOverride === "desktop") {
    return false;
  }
  return input.isMobilePlatform || input.compactViewport;
}
