<script>
export default {
  name: "TagStatus",
  props: ["status"],
  computed: {
    verbose() {
      return {
        pub: "Published",
        sked: "Scheduled",
        none: "Unpublished",
      }[this.status];
    },
    cssClass() {
      return {
        pub: "is-success",
        sked: "is-warning",
        none: "is-danger",
      }[this.status];
    },
  },
};
</script>

<template>
  <span class="tag is-small has-text-weight-semibold" :class="cssClass">
    <span class="icon is-size-6">
      <font-awesome-icon
        :icon="status === 'pub' ? ['fas', 'check-circle'] : ['fas', 'pen-nib']"
      ></font-awesome-icon>
    </span>
    <span>
      {{ verbose }}
    </span>
  </span>
</template>
