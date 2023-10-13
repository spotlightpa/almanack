<script setup>
import { post, postDonorWall } from "@/api/client-v2.js";
import { makeState } from "@/api/service-util.js";

const { exec, apiStateRefs } = makeState();
const isLoading = apiStateRefs.isLoadingThrottled;
const { error } = apiStateRefs;

function updateDonorWall() {
  return exec(() => post(postDonorWall, ""));
}
</script>

<template>
  <MetaHead>
    <title>Donor Walls • Spotlight PA Almanack</title>
  </MetaHead>

  <div class="px-2">
    <BulmaBreadcrumbs
      :links="[
        { name: 'Admin', to: { name: 'admin' } },
        { name: 'Donor Walls', to: { name: 'donor-wall' } },
      ]"
    />
    <h1 class="title">Donor Walls</h1>
  </div>

  <LinkButtons label="Links">
    <LinkHref
      :icon="['fas', 'table-list']"
      target="_blank"
      label="Google Sheet"
      href="/ssr/donor-wall"
    />
    <LinkHref
      target="_blank"
      label="Support Spotlight PA"
      href="https://www.spotlightpa.org/support/"
    />
    <LinkHref
      :icon="['fas', 'receipt']"
      label="Major Donors"
      href="https://www.spotlightpa.org/support/funders-and-members#major-donors-and-funders-since-launch"
      target="_blank"
    />
    <LinkHref
      :icon="['fas', 'receipt']"
      label="Members since launch"
      href="https://www.spotlightpa.org/support/funders-and-members#all-donors"
      target="_blank"
    />
    <LinkHref
      :icon="['fas', 'receipt']"
      label="Leaders in Action"
      href="https://www.spotlightpa.org/support/leaders-in-action/#our-current-leaders"
      target="_blank"
    />
    <LinkHref
      :icon="['fas', 'receipt']"
      label="State College"
      href="https://www.spotlightpa.org/support/state-college/#state-college-bureau-donors"
      target="_blank"
    />
  </LinkButtons>

  <div class="mt-5 mb-0 buttons">
    <button
      class="button has-text-weight-semibold is-success"
      :class="{ 'is-loading': isLoading }"
      type="button"
      @click="updateDonorWall"
    >
      Update donor walls from Google Sheet
    </button>
  </div>

  <ErrorSimple :error="error" />

  <p class="help">
    Allow five minutes for the live site to refresh with changes.
  </p>
</template>
