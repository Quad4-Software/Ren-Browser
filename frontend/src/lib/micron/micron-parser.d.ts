// SPDX-License-Identifier: MIT
declare module "micron-parser" {
  type MicronMultilineOptions = {
    windowMs?: number;
    rows?: number;
  };

  const BaseMicronParser: {
    enableDoubleEnterMultiline(
      container: HTMLElement,
      options?: MicronMultilineOptions,
    ): () => void;
  };

  export default BaseMicronParser;
}
