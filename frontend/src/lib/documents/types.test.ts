// SPDX-License-Identifier: MIT
import { describe, expect, it } from "vitest";
import {
  canonicalDocumentURL,
  decodeBase64ToUint8Array,
  documentURL,
  isDocumentContentType,
  isDocumentURL,
  isReadableDocumentName,
  isReadableMeshFileURL,
  parseDocumentPathFromURL,
  resolveDocumentAbsolutePath,
} from "./types";

describe("document types", () => {
  it("detects readable extensions", () => {
    expect(isReadableDocumentName("book.pdf")).toBe(true);
    expect(isReadableDocumentName("novel.EPUB")).toBe(true);
    expect(isReadableDocumentName("archive.zip")).toBe(false);
  });

  it("builds and recognizes document URLs", () => {
    const downloadDir = "/home/user/Downloads";
    const url = documentURL("/home/user/Downloads/book.pdf", downloadDir);
    expect(url).toBe("document:/book.pdf");
    expect(isDocumentURL(url)).toBe(true);
    expect(parseDocumentPathFromURL(url)).toBe("book.pdf");
    expect(resolveDocumentAbsolutePath(url, downloadDir)).toBe("/home/user/Downloads/book.pdf");
  });

  it("keeps legacy query URLs for paths outside the download folder", () => {
    const url = documentURL("/tmp/book.pdf", "/home/user/Downloads");
    expect(url).toBe("document:?path=%2Ftmp%2Fbook.pdf");
    expect(parseDocumentPathFromURL(url)).toBe("/tmp/book.pdf");
  });

  it("canonicalizes legacy document URLs", () => {
    const legacy = "document:?path=%2Fhome%2Fuser%2FDownloads%2Fbook.epub";
    const downloadDir = "/home/user/Downloads";
    expect(canonicalDocumentURL(legacy, downloadDir)).toBe("document:/book.epub");
  });

  it("detects readable mesh file links", () => {
    expect(isReadableMeshFileURL("deadbeef:/file/manual.pdf")).toBe(true);
    expect(isReadableMeshFileURL("deadbeef:/file/data.bin")).toBe(false);
    expect(isReadableMeshFileURL("document:/manual.pdf")).toBe(false);
    expect(isReadableMeshFileURL("document:?path=%2Fhome%2Fuser%2Fbook.epub")).toBe(false);
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
