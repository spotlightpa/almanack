import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import SiteParamsBanner from "@/components/SiteParamsBanner.vue";

describe("SiteParamsBanner", () => {
  const defaultParams = {
    data: {
      "banner-active": false,
      banner: "Test banner text",
      "banner-link": "/test-link",
      "banner-bg-color": "#ff6c36",
      "banner-text-color": "#ffffff",
    },
  };

  const mountComponent = (params = defaultParams) => {
    return mount(SiteParamsBanner, {
      props: {
        params,
        fileProps: {},
      },
      global: {
        stubs: {
          BulmaField: {
            template: '<div class="bulma-field"><slot></slot></div>',
            props: ["label", "help"],
          },
          BulmaTextarea: true,
          BulmaFieldInput: true,
          BulmaFieldColor: true,
        },
      },
    });
  };

  it("exposes saveData function via defineExpose", () => {
    const wrapper = mountComponent();
    expect(wrapper.vm.saveData).toBeDefined();
    expect(typeof wrapper.vm.saveData).toBe("function");
  });

  it("saveData returns data with correct keys for backend", () => {
    const wrapper = mountComponent();
    const saved = wrapper.vm.saveData();

    // These keys must match what the backend expects
    expect(saved).toHaveProperty("banner-active");
    expect(saved).toHaveProperty("banner");
    expect(saved).toHaveProperty("banner-link");
    expect(saved).toHaveProperty("banner-bg-color");
    expect(saved).toHaveProperty("banner-text-color");
  });

  it("saveData serializes link with toRel transform", () => {
    const params = {
      data: {
        ...defaultParams.data,
        "banner-link": "https://www.spotlightpa.org/news/article",
      },
    };
    const wrapper = mountComponent(params);
    const saved = wrapper.vm.saveData();

    // toRel should strip the domain
    expect(saved["banner-link"]).toBe("/news/article");
  });

  it("checkbox updates banner-active in saveData", async () => {
    const wrapper = mountComponent();
    const checkbox = wrapper.find('input[type="checkbox"]');

    expect(wrapper.vm.saveData()["banner-active"]).toBe(false);

    await checkbox.setValue(true);

    expect(wrapper.vm.saveData()["banner-active"]).toBe(true);
  });
});
