// SPDX-License-Identifier: MIT
import JSZip from "jszip";
import { describe, expect, it } from "vitest";
import { parseEpub } from "./epub";

async function buildNamespacedEpub(): Promise<Uint8Array> {
  const zip = new JSZip();
  zip.file(
    "META-INF/container.xml",
    `<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`,
  );
  zip.file(
    "OEBPS/content.opf",
    `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="uid">
  <metadata>
    <dc:title xmlns:dc="http://purl.org/dc/elements/1.1/">Test Book</dc:title>
  </metadata>
  <manifest>
    <item id="c1" href="chapter.xhtml" media-type="application/xhtml+xml"/>
  </manifest>
  <spine>
    <itemref idref="c1"/>
  </spine>
</package>`,
  );
  zip.file(
    "OEBPS/chapter.xhtml",
    `<?xml version="1.0"?>
<html xmlns="http://www.w3.org/1999/xhtml">
  <head><title>Chapter One</title></head>
  <body><h1>Chapter One</h1><p>Hello epub.</p></body>
</html>`,
  );
  return await zip.generateAsync({ type: "uint8array" });
}

describe("parseEpub", () => {
  it("parses namespaced OPF packages", async () => {
    const data = await buildNamespacedEpub();
    const book = await parseEpub(data);
    expect(book.title).toBe("Test Book");
    expect(book.chapters).toHaveLength(1);
    expect(book.chapters[0]?.title).toBe("Chapter One");
    expect(book.chapters[0]?.html).toContain("Hello epub.");
  });

  it("drops external image references before sanitizing", async () => {
    const zip = new JSZip();
    zip.file(
      "META-INF/container.xml",
      `<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`,
    );
    zip.file(
      "OEBPS/content.opf",
      `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="uid">
  <metadata><dc:title xmlns:dc="http://purl.org/dc/elements/1.1/">Unsafe</dc:title></metadata>
  <manifest><item id="c1" href="chapter.xhtml" media-type="application/xhtml+xml"/></manifest>
  <spine><itemref idref="c1"/></spine>
</package>`,
    );
    zip.file(
      "OEBPS/chapter.xhtml",
      `<?xml version="1.0"?>
<html xmlns="http://www.w3.org/1999/xhtml">
  <body>
    <p>Safe</p>
    <script>alert(1)</script>
    <img src="https://evil.test/x.png" alt="bad">
    <p style="background-image:url(https://evil.test/bg.png)">styled</p>
  </body>
</html>`,
    );
    const book = await parseEpub(await zip.generateAsync({ type: "uint8array" }));
    const html = book.chapters[0]?.html ?? "";
    expect(html).toContain("Safe");
    expect(html).not.toContain("https://evil.test");
    expect(html).not.toContain('src="https://');
  });

  it("uses NCX labels instead of repeated document titles", async () => {
    const zip = new JSZip();
    zip.file(
      "META-INF/container.xml",
      `<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`,
    );
    zip.file(
      "OEBPS/content.opf",
      `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="2.0" unique-identifier="uid">
  <metadata>
    <dc:title xmlns:dc="http://purl.org/dc/elements/1.1/">Repeated Book Title</dc:title>
  </metadata>
  <manifest>
    <item id="ncx" href="toc.ncx" media-type="application/x-dtbncx+xml"/>
    <item id="c1" href="chapter.xhtml" media-type="application/xhtml+xml"/>
  </manifest>
  <spine toc="ncx">
    <itemref idref="c1"/>
  </spine>
</package>`,
    );
    zip.file(
      "OEBPS/toc.ncx",
      `<?xml version="1.0"?>
<ncx xmlns="http://www.daisy.org/z3986/2005/ncx/" version="2005-1">
  <navMap>
    <navPoint id="c1">
      <navLabel><text>Chapter One</text></navLabel>
      <content src="chapter.xhtml"/>
    </navPoint>
  </navMap>
</ncx>`,
    );
    zip.file(
      "OEBPS/chapter.xhtml",
      `<?xml version="1.0"?>
<html xmlns="http://www.w3.org/1999/xhtml">
  <head><title>Repeated Book Title</title></head>
  <body><p>Hello epub.</p></body>
</html>`,
    );
    const book = await parseEpub(await zip.generateAsync({ type: "uint8array" }));
    expect(book.chapters[0]?.title).toBe("Chapter One");
  });
});
