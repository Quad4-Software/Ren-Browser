<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { AlertDialog } from "bits-ui";
  import Toggle from "$lib/components/Toggle.svelte";
  import PluginSignatureBadge from "$lib/components/PluginSignatureBadge.svelte";
  import { t, tPermission } from "$lib/i18n/i18n.svelte";

  export type PluginSignatureInfo = {
    present: boolean;
    valid: boolean;
    signer?: string;
    signerName?: string;
    trusted: boolean;
    error?: string;
  };

  export type PluginSecurityFinding = {
    id: string;
    severity: string;
    message: string;
  };

  export type PluginSecurityAssessment = {
    riskLevel: string;
    score: number;
    findings: PluginSecurityFinding[];
  };

  export type PluginInstallPreview = {
    id: string;
    name: string;
    version: string;
    description?: string;
    permissions?: string[];
    networkEndpoints: string[];
    requiresNetworkFetch: boolean;
    signature: PluginSignatureInfo;
    security?: PluginSecurityAssessment;
    i18nLocales?: string[];
  };

  export type PluginInstallChoices = {
    dontShowAgain: boolean;
    trustPublisher: boolean;
    grantedPermissions: string[];
  };

  type Props = {
    open: boolean;
    preview: PluginInstallPreview | null;
    confirming?: boolean;
    onConfirm: (choices: PluginInstallChoices) => void;
    onCancel: () => void;
  };

  let { open, preview, confirming = false, onConfirm, onCancel }: Props = $props();

  let confirmed = false;
  let dontShowAgain = $state(false);
  let trustPublisher = $state(false);
  const signature = $derived(
    preview?.signature ?? { present: false, valid: false, trusted: false },
  );
  let grantedPermissions = $state<Record<string, boolean>>({});

  $effect(() => {
    if (!open || !preview) {
      return;
    }
    confirmed = false;
    dontShowAgain = false;
    trustPublisher = false;
    const next: Record<string, boolean> = {};
    for (const perm of preview.permissions ?? []) {
      next[perm] = true;
    }
    grantedPermissions = next;
  });

  function permissionLabel(perm: string): string {
    return tPermission(perm);
  }

  function signatureLabel(info: PluginSignatureInfo): string {
    if (!info.present) {
      return t("extensions.signatureUnsigned");
    }
    if (!info.valid) {
      return t("extensions.signatureInvalid");
    }
    if (info.trusted && info.signerName) {
      return t("extensions.signatureTrusted", { name: info.signerName });
    }
    return t("extensions.signatureValid", { signer: info.signer ?? "" });
  }

  function selectedPermissions(): string[] {
    if (!preview?.permissions?.length) {
      return [];
    }
    return preview.permissions.filter((perm) => grantedPermissions[perm]);
  }

  function networkFetchGranted(): boolean {
    return grantedPermissions["network.fetch"] === true;
  }
</script>

<AlertDialog.Root
  open={open && preview !== null}
  onOpenChange={(next) => {
    if (!next) {
      if (!confirmed) {
        onCancel();
      }
      confirmed = false;
    }
  }}
>
  <AlertDialog.Portal>
    {#if preview}
      <AlertDialog.Overlay class="plugin-install-overlay" />
      <AlertDialog.Content
        class="plugin-install-dialog"
        interactOutsideBehavior={confirming ? "ignore" : "close"}
        onEscapeKeydown={(e) => {
          if (confirming) {
            e.preventDefault();
          }
        }}
      >
        <AlertDialog.Title id="plugin-network-install-title" class="plugin-install-title">
          {preview.requiresNetworkFetch
            ? t("extensions.networkInstallTitle")
            : t("extensions.installTitle")}
        </AlertDialog.Title>
        <AlertDialog.Description id="plugin-network-install-message" class="plugin-install-message">
          {t("extensions.installMessage", {
            name: preview.name,
            id: preview.id,
          })}
        </AlertDialog.Description>

        <div class="plugin-meta">
          <div class="title-row">
            <span class="version">v{preview.version}</span>
            <PluginSignatureBadge {signature} />
          </div>
          {#if preview.description}
            <p class="description">{preview.description}</p>
          {/if}
        </div>

        {#if preview.permissions?.length}
          <section class="permissions" aria-labelledby="plugin-permissions-heading">
            <h3 id="plugin-permissions-heading">{t("extensions.installPermissions")}</h3>
            <p class="muted">{t("extensions.installPermissionsHint")}</p>
            <ul class="permission-list">
              {#each preview.permissions as perm (perm)}
                <li>
                  <Toggle
                    label={permissionLabel(perm)}
                    checked={grantedPermissions[perm] ?? false}
                    onchange={(value) => {
                      grantedPermissions = { ...grantedPermissions, [perm]: value };
                    }}
                  />
                </li>
              {/each}
            </ul>
          </section>
        {/if}

        {#if preview.requiresNetworkFetch}
          <section class="endpoints" aria-labelledby="plugin-network-endpoints-heading">
            <h3 id="plugin-network-endpoints-heading">{t("extensions.networkInstallEndpoints")}</h3>
            {#if !networkFetchGranted()}
              <p class="muted">{t("extensions.networkInstallEndpointsBlocked")}</p>
            {/if}
            {#if (preview.networkEndpoints ?? []).length > 0}
              <ul>
                {#each preview.networkEndpoints ?? [] as endpoint (endpoint)}
                  <li><code>{endpoint}</code></li>
                {/each}
              </ul>
            {:else}
              <p class="muted">{t("extensions.networkInstallEndpointsUnknown")}</p>
            {/if}
          </section>
        {/if}

        {#if (preview.i18nLocales ?? []).length > 0}
          <section class="locales" aria-labelledby="plugin-i18n-locales-heading">
            <h3 id="plugin-i18n-locales-heading">{t("extensions.installI18nLocales")}</h3>
            <p class="locale-list">{(preview.i18nLocales ?? []).join(", ")}</p>
          </section>
        {/if}

        <section class="signature" aria-labelledby="plugin-signature-heading">
          <h3 id="plugin-signature-heading">{t("extensions.signatureTitle")}</h3>
          <p
            class:muted={!signature.present}
            class:error={signature.present && !signature.valid}
            class:trusted={signature.present && signature.valid && signature.trusted}
          >
            {signatureLabel(signature)}
          </p>
          {#if signature.present && signature.signer}
            <p class="signer"><code>{signature.signer}</code></p>
          {/if}
          {#if signature.present && !signature.valid && signature.error}
            <p class="error-detail">{signature.error}</p>
          {/if}
          {#if signature.present && signature.valid && !signature.trusted}
            <Toggle
              label={t("extensions.trustPublisher")}
              checked={trustPublisher}
              onchange={(value) => {
                trustPublisher = value;
              }}
            />
          {/if}
        </section>

        {#if preview.security?.findings?.length}
          <section class="security" aria-labelledby="plugin-security-heading">
            <h3 id="plugin-security-heading">
              {t("extensions.securityTitle", { level: preview.security.riskLevel })}
            </h3>
            <ul class="security-findings">
              {#each preview.security.findings as finding (finding.id)}
                <li data-severity={finding.severity}>{finding.message}</li>
              {/each}
            </ul>
          </section>
        {/if}

        <Toggle
          label={t("extensions.networkInstallDontShowAgain")}
          checked={dontShowAgain}
          onchange={(value) => {
            dontShowAgain = value;
          }}
        />

        <div class="actions">
          <AlertDialog.Cancel type="button" class="plugin-install-cancel-btn" disabled={confirming}>
            {t("common.cancel")}
          </AlertDialog.Cancel>
          <AlertDialog.Action
            type="button"
            class="plugin-install-confirm-btn"
            disabled={confirming || (signature.present && !signature.valid)}
            onclick={() => {
              confirmed = true;
              onConfirm({
                dontShowAgain,
                trustPublisher,
                grantedPermissions: selectedPermissions(),
              });
            }}
          >
            {t("extensions.networkInstallConfirm")}
          </AlertDialog.Action>
        </div>
      </AlertDialog.Content>
    {/if}
  </AlertDialog.Portal>
</AlertDialog.Root>

<style>
  :global(.plugin-install-overlay) {
    position: fixed;
    inset: 0;
    z-index: 1200;
    border: none;
    background: rgb(0 0 0 / 0.45);
    cursor: default;
  }

  :global(.plugin-install-dialog) {
    position: fixed;
    top: 50%;
    left: 50%;
    z-index: 1201;
    width: min(32rem, calc(100vw - 2rem));
    max-height: min(85vh, 40rem);
    overflow: auto;
    transform: translate(-50%, -50%);
    padding: 1.1rem 1.15rem 1rem;
    border: 1px solid var(--ren-border);
    border-radius: calc(var(--ren-radius) + 2px);
    background: var(--ren-chrome-bg);
    box-shadow: var(--ren-shadow);
    display: grid;
    gap: 0.85rem;
  }

  :global(.plugin-install-title) {
    margin: 0;
    font-size: 1rem;
    font-weight: 600;
    color: var(--ren-fg);
  }

  :global(.plugin-install-message) {
    margin: 0;
    font-size: 0.92rem;
    line-height: 1.45;
    color: var(--ren-fg-secondary);
  }

  h3 {
    margin: 0 0 0.35rem;
    font-size: 0.88rem;
    font-weight: 600;
    color: var(--ren-fg);
  }

  p {
    margin: 0;
    font-size: 0.92rem;
    line-height: 1.45;
    color: var(--ren-fg-secondary);
  }

  .plugin-meta {
    display: grid;
    gap: 0.25rem;
  }

  .title-row {
    display: flex;
    flex-wrap: wrap;
    gap: 0.45rem;
    align-items: center;
  }

  .version {
    color: var(--ren-muted);
    font-size: 0.82rem;
  }

  .description {
    font-size: 0.85rem;
  }

  .permissions {
    display: grid;
    gap: 0.45rem;
  }

  .permission-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    gap: 0.45rem;
  }

  .permission-list li {
    border: 1px solid var(--ren-border);
    border-radius: 8px;
    padding: 0.45rem 0.55rem;
  }

  .endpoints ul {
    margin: 0;
    padding-left: 1.1rem;
    display: grid;
    gap: 0.35rem;
  }

  .locale-list {
    margin: 0;
    color: var(--ren-muted);
    font-size: 0.85rem;
  }

  .endpoints li {
    font-size: 0.85rem;
    color: var(--ren-fg-secondary);
    overflow-wrap: anywhere;
    word-break: break-word;
  }

  code {
    font-family: var(--ren-mono, monospace);
    font-size: 0.82rem;
  }

  .muted {
    color: var(--ren-muted);
    font-size: 0.85rem;
  }

  .signature {
    display: grid;
    gap: 0.45rem;
  }

  .signature .trusted {
    color: var(--ren-fg-secondary);
  }

  .signature .error,
  .error-detail {
    color: var(--ren-danger, #e5484d);
    font-size: 0.85rem;
    overflow-wrap: anywhere;
    word-break: break-word;
  }

  .signer {
    margin: 0;
    font-size: 0.82rem;
    overflow-wrap: anywhere;
    word-break: break-all;
  }

  .security-findings {
    margin: 0;
    padding-left: 1.1rem;
    display: grid;
    gap: 0.3rem;
    font-size: 0.82rem;
  }

  .security-findings li[data-severity="high"] {
    color: var(--ren-danger, #e5484d);
  }

  .security-findings li[data-severity="warn"] {
    color: #f0c674;
  }

  .actions {
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
    padding-top: 0.15rem;
  }

  :global(.plugin-install-cancel-btn),
  :global(.plugin-install-confirm-btn) {
    border: 1px solid var(--ren-border);
    border-radius: 10px;
    padding: 0.5rem 0.85rem;
    font: inherit;
    font-size: 0.88rem;
    cursor: pointer;
    transition:
      background 0.15s ease,
      border-color 0.15s ease,
      color 0.15s ease;
  }

  :global(.plugin-install-cancel-btn) {
    background: transparent;
    color: var(--ren-fg);
  }

  :global(.plugin-install-cancel-btn:hover:not(:disabled)) {
    background: var(--ren-tab-hover);
  }

  :global(.plugin-install-cancel-btn:disabled),
  :global(.plugin-install-confirm-btn:disabled) {
    opacity: 0.55;
    cursor: not-allowed;
  }

  :global(.plugin-install-confirm-btn) {
    background: var(--ren-accent);
    border-color: var(--ren-accent);
    color: #fff;
  }

  :global(.plugin-install-confirm-btn:hover:not(:disabled)) {
    background: var(--ren-accent-hover);
    border-color: var(--ren-accent-hover);
  }
</style>
