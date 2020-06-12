<template>
  <div
    class="textarea pre-wrap no-select has-margin-bottom"
    rows="bulmaoverride"
  >
    <slot />
  </div>
</template>

<style scoped>
.no-select {
  user-select: none;
  cursor: not-allowed;
}
</style>
