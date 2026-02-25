import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import TagStatus from "@/components/TagStatus.vue";

describe("TagStatus", () => {
  it('renders "Published" for pub status with success class', () => {
    const wrapper = mount(TagStatus, {
      props: { status: "pub" },
      global: {
        stubs: { "font-awesome-icon": true },
      },
    });
    expect(wrapper.text()).toBe("Published");
    expect(wrapper.classes()).toContain("is-success");
  });

  it('renders "Scheduled" for sked status with warning class', () => {
    const wrapper = mount(TagStatus, {
      props: { status: "sked" },
      global: {
        stubs: { "font-awesome-icon": true },
      },
    });
    expect(wrapper.text()).toBe("Scheduled");
    expect(wrapper.classes()).toContain("is-warning");
  });

  it('renders "Unpublished" for none status with danger class', () => {
    const wrapper = mount(TagStatus, {
      props: { status: "none" },
      global: {
        stubs: { "font-awesome-icon": true },
      },
    });
    expect(wrapper.text()).toBe("Unpublished");
    expect(wrapper.classes()).toContain("is-danger");
  });
});
