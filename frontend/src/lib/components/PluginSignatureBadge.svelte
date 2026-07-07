<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { t } from "$lib/i18n/i18n.svelte";

  export type PluginSignatureInfo = {
    present: boolean;
    valid: boolean;
    signer?: string;
    signerName?: string;
    trusted: boolean;
    error?: string;
  };

  type Props = {
    signature: PluginSignatureInfo;
    tampered?: boolean;
    compact?: boolean;
  };

  let { signature, tampered = false, compact = false }: Props = $props();

  type BadgeKind = "unsigned" | "signed" | "trusted" | "invalid";

  const badgeKind = $derived.by((): BadgeKind => {
    if (!signature.present) {
      return "unsigned";
    }
    if (!signature.valid) {
      return "invalid";
    }
    if (signature.trusted) {
      return "trusted";
    }
    return "signed";
  });

  function badgeLabel(kind: BadgeKind): string {
    switch (kind) {
      case "unsigned":
        return t("extensions.badgeUnsigned");
      case "signed":
        return t("extensions.badgeSigned");
      case "trusted":
        return t("extensions.badgeTrusted", { name: signature.signerName ?? "" });
      case "invalid":
        return t("extensions.badgeInvalid");
    }
  }
</script>

<span class="badges" class:compact>
  {#if tampered}
    <span class="badge tampered" title={t("extensions.badgeTamperedHint")}>
      {t("extensions.badgeTampered")}
    </span>
  {/if}
  <span class="badge {badgeKind}" title={badgeLabel(badgeKind)}>
    {badgeLabel(badgeKind)}
  </span>
</span>

<style>
  .badges {
    display: inline-flex;
    flex-wrap: wrap;
    gap: 0.35rem;
    align-items: center;
  }

  .badge {
    display: inline-flex;
    align-items: center;
    border-radius: 999px;
    padding: 0.12rem 0.5rem;
    font-size: 0.68rem;
    font-weight: 600;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    border: 1px solid transparent;
    white-space: nowrap;
  }

  .compact .badge {
    font-size: 0.62rem;
    padding: 0.08rem 0.4rem;
  }

  .unsigned {
    color: var(--ren-muted);
    border-color: var(--ren-border);
    background: color-mix(in srgb, var(--ren-muted) 12%, transparent);
  }

  .signed {
    color: #8ec8ff;
    border-color: color-mix(in srgb, #8ec8ff 45%, transparent);
    background: color-mix(in srgb, #8ec8ff 12%, transparent);
  }

  .trusted {
    color: #7fdca0;
    border-color: color-mix(in srgb, #7fdca0 45%, transparent);
    background: color-mix(in srgb, #7fdca0 12%, transparent);
  }

  .invalid,
  .tampered {
    color: #ff9b9b;
    border-color: color-mix(in srgb, #ff9b9b 45%, transparent);
    background: color-mix(in srgb, #ff9b9b 12%, transparent);
  }
</style>
