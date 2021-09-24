<script>
export default {
  name: "APILoader",
  props: {
    isLoading: Boolean,
    reload: Function,
    error: Error,
  },
};
</script>

<template>
  <div>
    <SpinnerProgress :is-loading="isLoading" />
    <ErrorReloader :error="error" @reload="reload" />
    <div v-if="!isLoading && !error">
      <slot></slot>
    </div>
  </div>
</template>
