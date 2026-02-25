import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import { defineComponent } from "vue";
import SiteParams from "@/components/SiteParams.vue";

// Create mock child components that expose saveData
const createMockChild = (name, data) =>
  defineComponent({
    name,
    props: ["params", "fileProps"],
    expose: ["saveData"],
    methods: {
      saveData() {
        return data;
      },
    },
    template: `<div class="mock-${name}"></div>`,
  });

describe("SiteParams", () => {
  const createStubs = (overrides = {}) => ({
    SiteParamsBanner: createMockChild("banner", { "banner-active": true }),
    SiteParamsHeader: createMockChild("header", { "header-slogan": "test" }),
    SiteParamsRailTop: createMockChild("railTop", { "rail-top": true }),
    SiteParamsRailSticky: createMockChild("railSticky", { "rail-sticky": true }),
    SiteParamsFooter: createMockChild("footer", { "footer-cta": "test" }),
    SiteParamsFeaturedHomepage: createMockChild("featuredHomepage", {
      "homepage-active": true,
    }),
    SiteParamsBreaker: createMockChild("breaker", { "breaker-active": false }),
    SiteParamsHeadwater: createMockChild("headwater", { headwater: "hw" }),
    SiteParamsRiver: createMockChild("river", { river: "content" }),
    SiteParamsSupport: createMockChild("support", { "support-active": true }),
    SiteParamsFeaturedArticle: createMockChild("featuredArticle", {
      "featured-active": true,
    }),
    SiteParamsSticky: createMockChild("sticky", { "sticky-active": true }),
    SiteParamsNewsletter: createMockChild("newsletter", {
      "newsletter-active": true,
    }),
    SiteParamsTakeover: createMockChild("takeover", { "takeover-active": false }),
    ...overrides,
  });

  it("saveParams aggregates saveData from all child refs", () => {
    const wrapper = mount(SiteParams, {
      props: {
        params: {
          scheduleFor: "2024-01-01T00:00:00Z",
          data: {
            existingKey: "existingValue",
          },
        },
        fileProps: {},
      },
      global: {
        stubs: createStubs(),
      },
    });

    const result = wrapper.vm.saveParams();

    // Should include schedule_for
    expect(result.schedule_for).toBe("2024-01-01T00:00:00Z");

    // Should include original data
    expect(result.data.existingKey).toBe("existingValue");

    // Should include data from child saveData calls
    expect(result.data["banner-active"]).toBe(true);
    expect(result.data["breaker-active"]).toBe(false);
    expect(result.data["featured-active"]).toBe(true);
    expect(result.data["newsletter-active"]).toBe(true);
    expect(result.data["sticky-active"]).toBe(true);
    expect(result.data["support-active"]).toBe(true);
    expect(result.data["takeover-active"]).toBe(false);
  });

  it("exposes saveParams via defineExpose", () => {
    const wrapper = mount(SiteParams, {
      props: {
        params: { scheduleFor: null, data: {} },
        fileProps: {},
      },
      global: {
        stubs: createStubs(),
      },
    });

    expect(wrapper.vm.saveParams).toBeDefined();
    expect(typeof wrapper.vm.saveParams).toBe("function");
  });
});
