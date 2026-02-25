import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import ErrorSimple from "@/components/ErrorSimple.vue";

describe("ErrorSimple", () => {
  it("renders error name and message from Error object", () => {
    const wrapper = mount(ErrorSimple, {
      props: { error: new Error("Something went wrong") },
    });
    expect(wrapper.text()).toContain("Error");
    expect(wrapper.text()).toContain("Something went wrong");
  });

  it("renders string error as both name and message", () => {
    const wrapper = mount(ErrorSimple, {
      props: { error: "Network error" },
    });
    expect(wrapper.text()).toContain("Network error");
  });

  it("does not render when error is null", () => {
    const wrapper = mount(ErrorSimple, {
      props: { error: null },
    });
    expect(wrapper.find(".message").exists()).toBe(false);
  });
});
