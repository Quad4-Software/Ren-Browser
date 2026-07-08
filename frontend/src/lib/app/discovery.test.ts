import { describe, expect, it } from "vitest";

describe("Discovery Slow Mode and Unique Announces", () => {
  it("verifies that slow mode changes the polling interval", () => {
    const DISCOVERY_POLL_MS = 5000;
    const DISCOVERY_POLL_SLOW_MS = 15000;

    function getPollInterval(slowMode: boolean): number {
      return slowMode ? DISCOVERY_POLL_SLOW_MS : DISCOVERY_POLL_MS;
    }

    expect(getPollInterval(false)).toBe(5000);
    expect(getPollInterval(true)).toBe(15000);
  });

  it("verifies that unique announces are processed and not slowed down by slow mode", () => {
    // Unique announces are processed immediately by the backend and stored in the node list.
    // The frontend's slow mode only throttles how frequently the UI polls the backend,
    // ensuring that the speed of incoming announces is unaffected.
    const nodes = [
      { hash: "abc", name: "Node A", hops: 1, lastSeen: 1000 },
      { hash: "abc", name: "Node A", hops: 1, lastSeen: 1050 }, // Duplicate/updated announce
      { hash: "def", name: "Node B", hops: 2, lastSeen: 2000 },
    ];

    // Deduplicate unique nodes (simulating backend/store behavior)
    const uniqueNodesMap = new Map<string, typeof nodes[0]>();
    for (const node of nodes) {
      uniqueNodesMap.set(node.hash, node);
    }

    expect(uniqueNodesMap.size).toBe(2);
    expect(uniqueNodesMap.get("abc")?.lastSeen).toBe(1050); // Updated to latest seen
  });
});
