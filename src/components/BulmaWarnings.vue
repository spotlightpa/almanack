<script>
export default {
  props: {
    values: {
      type: Array,
      required: true,
    },
  },
  computed: {
    active() {
      return this.values
        .filter(([cond]) => cond)
        .map(([, link, msg]) => ({ link, msg }));
    },
  },
  methods: {
    scrollTo(link) {
      document.querySelector(link)?.scrollIntoView({
        behavior: "smooth",
        block: "start",
      });
    },
  },
};
</script>

<template>
  <div v-if="active.length" class="message is-warning">
    <div class="message-header">
      <span>
        <font-awesome-icon
          :icon="['fas', 'circle-exclamation']"
        ></font-awesome-icon>

        <span
          class="ml-1"
          v-text="active.length === 1 ? 'Warning' : 'Warnings'"
        ></span>
      </span>
    </div>
    <div class="message-body">
      <li v-for="{ link, msg } of active" :key="msg">
        <a @click.prevent="scrollTo(link)">
          {{ msg }}
        </a>
      </li>
    </div>
  </div>
</template>

<style>
html {
  scroll-behavior: smooth;
}
</style>
