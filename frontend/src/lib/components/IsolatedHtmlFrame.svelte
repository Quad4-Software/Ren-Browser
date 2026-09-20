<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { useEventListener, useMutationObserver } from "runed";
  import { buildIsolatedHtmlDocument, ISOLATED_FRAME_SANDBOX } from "$lib/documents/isolated-html";
  import { resolvedReaderTheme, type ReaderTheme } from "$lib/documents/reader-theme";

  type Props = {
    html: string;
    title: string;
    fontScale?: number;
    rotation?: number;
    onScrollRoot?: (root: HTMLElement | null) => void;
    onFrame?: (frame: HTMLIFrameElement | null) => void;
  };

  let {
    html,
    title,
    fontScale = 1,
    rotation = 0,
    onScrollRoot = () => {},
    onFrame = () => {},
  }: Props = $props();

  let frameEl: HTMLIFrameElement | undefined = $state();
  let readerTheme = $state<ReaderTheme>(resolvedReaderTheme());

  useMutationObserver(
    () => document.documentElement,
    () => {
      readerTheme = resolvedReaderTheme();
    },
    { attributes: true, attributeFilter: ["data-theme"] },
  );

  function scrollRootFromFrame(frame: HTMLIFrameElement | undefined): HTMLElement | null {
    const doc = frame?.contentDocument;
    if (!doc) {
      return null;
    }
    return doc.scrollingElement instanceof HTMLElement ? doc.scrollingElement : doc.documentElement;
  }

  function writeFrameContent(
    frame: HTMLIFrameElement,
    bodyHtml: string,
    theme: ReaderTheme,
    scale: number,
    rotate: number,
  ) {
    const doc = frame.contentDocument;
    if (!doc) {
      return;
    }
    doc.open();
    doc.write(buildIsolatedHtmlDocument(bodyHtml, { theme, fontScale: scale, rotation: rotate }));
    doc.close();
    onScrollRoot(scrollRootFromFrame(frame));
  }

  $effect(() => {
    onFrame(frameEl ?? null);
    return () => {
      onFrame(null);
    };
  });

  useEventListener(
    () => frameEl,
    "load",
    () => {
      const frame = frameEl;
      if (!frame || !html) {
        return;
      }
      writeFrameContent(frame, html, readerTheme, fontScale, rotation);
    },
    { once: true },
  );

  $effect(() => {
    const frame = frameEl;
    const bodyHtml = html;
    const theme = readerTheme;
    const scale = fontScale;
    const rotate = rotation;
    if (!frame || !bodyHtml) {
      onScrollRoot(null);
      return;
    }

    if (frame.contentDocument?.readyState === "complete") {
      writeFrameContent(frame, bodyHtml, theme, scale, rotate);
    }

    return () => {
      onScrollRoot(null);
    };
  });
</script>

<iframe
  bind:this={frameEl}
  class="isolated-frame"
  {title}
  sandbox={ISOLATED_FRAME_SANDBOX}
  src="about:blank"
></iframe>

<style>
  .isolated-frame {
    width: 100%;
    flex: 1;
    min-height: 0;
    border: none;
    display: block;
    background: transparent;
    color: inherit;
  }
</style>
