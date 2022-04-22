<script>
import Vue from "vue";
import { reactive, toRefs, watch } from "@vue/composition-api";

import { useClient, makeState } from "@/api/hooks.js";
import { useFileList } from "@/api/file-list.js";

import { formatDateTime } from "@/utils/time-format.js";
import sanitizeText from "@/utils/sanitize-text.js";

class SiteParams {
  constructor(config) {
    let scheduleFor = config.schedule_for;
    this.scheduleFor = scheduleFor ? new Date(scheduleFor) : null;
    let pub = config.published_at;
    this.publishedAt = pub ? new Date(pub) : null;
    this.isCurrent = !!this.publishedAt;

    let data = config.data;
    this.bannerActive = data["banner-active"] ?? false;
    this.bannerText = data["banner"] ?? "";
    this.bannerLink = SiteParams.link(data, "banner-link");
    this.bannerBgColor = data["banner-bg-color"] ?? "";
    this.bannerTextColor = data["banner-text-color"] ?? "";

    this.topperActive = data["topper-active"] ?? false;
    this.topperBgColor = data["topper-bg-color"] ?? "";
    this.topperDividerColor = data["topper-divider-color"] ?? "";
    this.topperLink = SiteParams.link(data, "topper-link");
    this.topperImageDescription = data["topper-image-description"] ?? "";
    this.topperDesktopHeight = data["topper-desktop-height"] ?? 0;
    this.topperDesktopWidth = data["topper-desktop-width"] ?? 0;
    this.topperDesktopImages = data["topper-desktop-images"] ?? [];
    this.topperMobileHeight = data["topper-mobile-height"] ?? 0;
    this.topperMobileWidth = data["topper-mobile-width"] ?? 0;
    this.topperMobileImages = data["topper-mobile-images"] ?? [];

    this.promoActive = data["promo-active"] ?? false;
    this.promoType = data["promo-type"] ?? "";
    this.promoImageDescription = data["promo-image-description"] ?? "";
    this.promoDesktopImages = data["promo-desktop-images"] ?? [];
    this.promoDesktopWidth = data["promo-desktop-width"] ?? 0;
    this.promoDesktopHeight = data["promo-desktop-height"] ?? 0;
    this.promoMobileImages = data["promo-mobile-images"] ?? [];
    this.promoMobileWidth = data["promo-mobile-width"] ?? 0;
    this.promoMobileHeight = data["promo-mobile-height"] ?? 0;
    this.promoLink = SiteParams.link(data, "promo-link");
    this.promoText = data["promo-text"] ?? "";

    this.stickyActive = data["sticky-active"] ?? false;
    this.stickyImageDescription = data["sticky-image-description"] ?? "";
    this.stickyImages = data["sticky-images"] ?? [];
    this.stickyLink = SiteParams.link(data, "sticky-link");

    this.riverDescription = data["river-promo-description"] ?? "";
    this.riverLink = SiteParams.link(data, "river-promo-link");
    this.riverDesktopImages = data["river-promo-desktop-images"] ?? [];
    this.riverDesktopWidth = data["river-promo-desktop-width"] ?? 0;
    this.riverDesktopHeight = data["river-promo-desktop-height"] ?? 0;
    this.riverMobileImages = data["river-promo-mobile-images"] ?? [];
    this.riverMobileWidth = data["river-promo-mobile-width"] ?? 0;
    this.riverMobileHeight = data["river-promo-mobile-height"] ?? 0;

    this.sidebarTopDescription = data["sidebar-top-description"] ?? "";
    this.sidebarTopLink = SiteParams.link(data, "sidebar-top-link");
    this.sidebarTopImages = data["sidebar-top-images"] ?? [];
    this.sidebarTopWidth = data["sidebar-top-width"] ?? 0;
    this.sidebarTopHeight = data["sidebar-top-height"] ?? 0;

    this.sidebarStickyDescription = data["sidebar-sticky-description"] ?? "";
    this.sidebarStickyLink = SiteParams.link(data, "sidebar-sticky-link");
    this.sidebarStickyImages = data["sidebar-sticky-images"] ?? [];
    this.sidebarStickyWidth = data["sidebar-sticky-width"] ?? 0;
    this.sidebarStickyHeight = data["sidebar-sticky-height"] ?? 0;

    this.footerDescription = data["footer-promo-description"] ?? "";
    this.footerLink = SiteParams.link(data, "footer-promo-link");
    this.footerDesktopImages = data["footer-promo-desktop-images"] ?? [];
    this.footerDesktopWidth = data["footer-promo-desktop-width"] ?? 0;
    this.footerDesktopHeight = data["footer-promo-desktop-height"] ?? 0;
    this.footerMobileImages = data["footer-promo-mobile-images"] ?? [];
    this.footerMobileWidth = data["footer-promo-mobile-width"] ?? 0;
    this.footerMobileHeight = data["footer-promo-mobile-height"] ?? 0;

    this.supportActive = data["support-active"] ?? false;
    this.supportLink = SiteParams.link(data, "support-link");
    this.supportHed = data["support-hed"] ?? "";
    this.supportText = data["support-text"] ?? "";
    this.supportCTA = data["support-cta"] ?? "";
    this.supportHedColor = data["support-hed-color"] ?? "";
    this.supportTextColor = data["support-text-color"] ?? "";
    this.supportBgColor = data["support-bg-color"] ?? "";
    this.supportButtonBgColor = data["support-button-bg-color"] ?? "";
    this.supportButtonTextColor = data["support-button-text-color"] ?? "";

    this.newsletterActive = data["newsletter-active"] ?? false;
    Vue.observable(this);
  }

  toJSON() {
    return {
      schedule_for: this.scheduleFor,
      data: {
        ["banner-active"]: this.bannerActive,
        ["banner"]: this.bannerHTML,
        ["banner-link"]: SiteParams.unlink(this.bannerLink),
        ["banner-bg-color"]: this.bannerBgColor,
        ["banner-text-color"]: this.bannerTextColor,
        ["topper-active"]: this.topperActive,
        ["topper-bg-color"]: this.topperBgColor,
        ["topper-divider-color"]: this.topperDividerColor,
        ["topper-link"]: SiteParams.unlink(this.topperLink),
        ["topper-image-description"]: this.topperImageDescription,
        ["topper-desktop-height"]: this.topperDesktopHeight,
        ["topper-desktop-width"]: this.topperDesktopWidth,
        ["topper-desktop-images"]: this.topperDesktopImages,
        ["topper-mobile-height"]: this.topperMobileHeight,
        ["topper-mobile-width"]: this.topperMobileWidth,
        ["topper-mobile-images"]: this.topperMobileImages,
        ["promo-active"]: this.promoActive,
        ["promo-type"]: this.promoType,
        ["promo-image-description"]: this.promoImageDescription,
        ["promo-desktop-images"]: this.promoDesktopImages,
        ["promo-desktop-width"]: this.promoDesktopWidth,
        ["promo-desktop-height"]: this.promoDesktopHeight,
        ["promo-mobile-images"]: this.promoMobileImages,
        ["promo-mobile-width"]: this.promoMobileWidth,
        ["promo-mobile-height"]: this.promoMobileHeight,
        ["promo-link"]: SiteParams.unlink(this.promoLink),
        ["promo-text"]: this.promoText,
        ["sticky-active"]: this.stickyActive,
        ["sticky-image-description"]: this.stickyImageDescription,
        ["sticky-images"]: this.stickyImages,
        ["sticky-link"]: SiteParams.unlink(this.stickyLink),
        ["newsletter-active"]: this.newsletterActive,
        ["river-promo-description"]: this.riverDescription,
        ["river-promo-link"]: SiteParams.unlink(this.riverLink),
        ["river-promo-desktop-images"]: this.riverDesktopImages,
        ["river-promo-desktop-width"]: this.riverDesktopWidth,
        ["river-promo-desktop-height"]: this.riverDesktopHeight,
        ["river-promo-mobile-images"]: this.riverMobileImages,
        ["river-promo-mobile-width"]: this.riverMobileWidth,
        ["river-promo-mobile-height"]: this.riverMobileHeight,
        ["sidebar-top-description"]: this.sidebarTopDescription,
        ["sidebar-top-link"]: SiteParams.unlink(this.sidebarTopLink),
        ["sidebar-top-images"]: this.sidebarTopImages,
        ["sidebar-top-width"]: this.sidebarTopWidth,
        ["sidebar-top-height"]: this.sidebarTopHeight,
        ["sidebar-sticky-description"]: this.sidebarStickyDescription,
        ["sidebar-sticky-link"]: SiteParams.unlink(this.sidebarStickyLink),
        ["sidebar-sticky-images"]: this.sidebarStickyImages,
        ["sidebar-sticky-width"]: this.sidebarStickyWidth,
        ["sidebar-sticky-height"]: this.sidebarStickyHeight,
        ["footer-promo-description"]: this.footerDescription,
        ["footer-promo-link"]: SiteParams.unlink(this.footerLink),
        ["footer-promo-desktop-images"]: this.footerDesktopImages,
        ["footer-promo-desktop-width"]: this.footerDesktopWidth,
        ["footer-promo-desktop-height"]: this.footerDesktopHeight,
        ["footer-promo-mobile-images"]: this.footerMobileImages,
        ["footer-promo-mobile-width"]: this.footerMobileWidth,
        ["footer-promo-mobile-height"]: this.footerMobileHeight,
        ["support-active"]: this.supportActive,
        ["support-link"]: SiteParams.unlink(this.supportLink),
        ["support-hed"]: this.supportHed,
        ["support-text"]: this.supportHTML,
        ["support-cta"]: this.supportCTA,
        ["support-hed-color"]: this.supportHedColor,
        ["support-text-color"]: this.supportTextColor,
        ["support-bg-color"]: this.supportBgColor,
        ["support-button-bg-color"]: this.supportButtonBgColor,
        ["support-button-text-color"]: this.supportButtonTextColor,
      },
    };
  }

  get bannerHTML() {
    return sanitizeText(this.bannerText);
  }

  get supportHTML() {
    return sanitizeText(this.supportText);
  }

  static link(data, key) {
    let link = data[key];
    if (!link) {
      return "";
    }
    return new URL(link, "https://www.spotlightpa.org").href;
  }

  static unlink(url) {
    if (!url) {
      return "";
    }
    let u = new URL(url);
    if (
      u.hostname === "www.spotlightpa.org" ||
      u.hostname === "spotlightpa.org"
    ) {
      return u.pathname;
    }
    return url;
  }
}

export default {
  metaInfo: {
    title: "Sitewide Settings",
  },
  setup() {
    let { getSiteParams, postSiteParams } = useClient();
    const { apiState, exec } = makeState();

    const state = reactive({
      ...toRefs(apiState),

      configs: [],
      nextSchedule: null,
    });

    let actions = {
      fetch() {
        return exec(() => getSiteParams());
      },
      save() {
        return exec(() =>
          postSiteParams({
            configs: state.configs,
          })
        );
      },
      init() {
        if (!apiState.rawData) {
          return;
        }
        state.configs = apiState.rawData.configs.map(
          (data) => new SiteParams(data)
        );
      },
      async addScheduledConfig() {
        let lastParams = state.configs[state.configs.length - 1];
        state.configs.push(
          new SiteParams({
            ...JSON.parse(JSON.stringify(lastParams)),
            schedule_for: state.nextSchedule,
          })
        );
        state.nextSchedule = null;
        await this.$nextTick();
        // TODO: Fix this array if we ever upgrade to Vue 3
        // https://vueuse.org/core/useTemplateRefsList/
        this.$refs.configEls[this.$refs.configEls.length - 1].scrollIntoView({
          behavior: "smooth",
          block: "start",
        });
      },
      removeScheduledConfig(i) {
        state.configs.splice(i, 1);
      },
    };

    watch(
      () => apiState.rawData,
      () => actions.init()
    );

    actions.fetch();

    return {
      ...toRefs(state),
      ...actions,

      formatDateTime,
      files: useFileList(),
    };
  },
};
</script>

<template>
  <div>
    <div class="px-2">
      <BulmaBreadcrumbs
        :links="[
          { name: 'Admin', to: { name: 'admin' } },
          { name: 'Sitewide Settings', to: { name: 'site-params' } },
        ]"
      />
      <h1 class="title">Sitewide Settings</h1>
    </div>

    <template v-if="configs.length">
      <div v-for="(params, i) of configs" :key="i" class="px-2 py-4 zebra-row">
        <h2 ref="configEls" class="title is-3">
          {{
            params.isCurrent
              ? "Current Settings"
              : `Scheduled for ${formatDateTime(params.scheduleFor)}`
          }}
        </h2>
        <SiteParams :params="params" :file-props="files" />
        <button
          v-if="!params.isCurrent"
          type="button"
          class="mt-2 button is-danger has-text-weight-semibold"
          @click="removeScheduledConfig(i)"
        >
          Remove {{ formatDateTime(params.scheduleFor) }}
        </button>
      </div>
      <h2 class="mt-2 title">Add a scheduled change</h2>
      <BulmaField v-slot="{ idForLabel }" label="Schedule for">
        <b-datetimepicker
          :id="idForLabel"
          v-model="nextSchedule"
          icon="user-clock"
          :datetime-formatter="formatDateTime"
          :inline="true"
          locale="en-US"
        />
        <button
          type="button"
          :disabled="!nextSchedule || nextSchedule < new Date()"
          class="mt-3 button is-success has-text-weight-semibold"
          @click="addScheduledConfig"
        >
          <span class="icon is-size-6">
            <font-awesome-icon :icon="['fas', 'plus']" />
          </span>
          <span>Add</span>
        </button>
      </BulmaField>
    </template>

    <div class="buttons">
      <button
        type="button"
        class="button is-primary has-text-weight-semibold"
        :disabled="isLoading"
        :class="{ 'is-loading': isLoadingThrottled }"
        @click="save"
      >
        Save
      </button>
      <button
        type="button"
        class="button is-light has-text-weight-semibold"
        :disabled="isLoading"
        :class="{ 'is-loading': isLoadingThrottled }"
        @click="fetch"
      >
        Revert
      </button>
    </div>

    <SpinnerProgress :is-loading="isLoadingThrottled" />
    <ErrorReloader :error="error" @reload="fetch" />
  </div>
</template>

<style scoped>
.zebra-row {
  background-color: #fff;
}

.zebra-row:nth-child(odd) {
  background-color: #fafafa;
}

.zebra-row + .zebra-row {
  border-top: 1px solid #dbdbdb;
}
</style>
