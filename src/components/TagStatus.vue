<script setup>
import { computed } from "vue";

const props = defineProps(["status"]);

const verbose = computed(
  () =>
    ({
      pub: "Published",
      sked: "Scheduled",
      none: "Unpublished",
    })[props.status]
);

const cssClass = computed(
  () =>
    ({
      pub: "is-success",
      sked: "is-warning",
      none: "is-danger",
    })[props.status]
);
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
