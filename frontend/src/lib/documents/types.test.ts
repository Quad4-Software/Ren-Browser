// SPDX-License-Identifier: MIT
import { describe, expect, it } from "vitest";
import {
  decodeBase64ToUint8Array,
  documentURL,
  isDocumentContentType,
  isDocumentURL,
  isReadableDocumentName,
  isReadableMeshFileURL,
} from "./types";

describe("document types", () => {
  it("detects readable extensions", () => {
    expect(isReadableDocumentName("book.pdf")).toBe(true);
    expect(isReadableDocumentName("novel.EPUB")).toBe(true);
    expect(isReadableDocumentName("archive.zip")).toBe(false);
  });

  it("builds and recognizes document URLs", () => {
    const url = documentURL("/home/user/Downloads/book.pdf");
    expect(isDocumentURL(url)).toBe(true);
    expect(new URL(url).searchParams.get("path")).toBe("/home/user/Downloads/book.pdf");
  });

  it("detects readable mesh file links", () => {
    expect(isReadableMeshFileURL("deadbeef:/file/manual.pdf")).toBe(true);
    expect(isReadableMeshFileURL("deadbeef:/file/data.bin")).toBe(false);
  });

  it("decodes base64 payloads", () => {
    const bytes = decodeBase64ToUint8Array("YQ==");
    expect(bytes).toEqual(new Uint8Array([97]));
  });

  it("recognizes document content types", () => {
    expect(isDocumentContentType("pdf")).toBe(true);
    expect(isDocumentContentType("epub")).toBe(true);
    expect(isDocumentContentType("html")).toBe(false);
  });
});
