<script>
export default {
  async mounted() {
    await this.$nextTick();
    this.$emit(
      "mounted",
      Array.from(this.$el.children)
        .map((el) =>
          el.tagName === "RAW-HTML" ? el.getAttribute("block") : el.outerHTML
        )
        .join("\n\n")
    );
  },
};
</script>

<template>
  <div hidden>
    <slot></slot>
  </div>
</template>
