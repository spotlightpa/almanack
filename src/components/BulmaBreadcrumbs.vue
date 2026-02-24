<script setup>
import { useRouter } from "vue-router";

defineProps({ links: Array });

const { resolve, currentRoute } = useRouter();

function isSelf({ to }) {
  let route = resolve(to);
  if (route.name !== currentRoute.value.name) {
    return false;
  }
  return (
    JSON.stringify(route.query) === JSON.stringify(currentRoute.value.query)
  );
}
</script>

<template>
  <nav class="breadcrumb has-succeeds-separator" aria-label="breadcrumbs">
    <ul>
      <li
        v-for="(link, i) of links"
        :key="i"
        :class="{ 'is-active': isSelf(link) }"
      >
        <router-link :to="link.to">
          {{ link.name }}
        </router-link>
      </li>
    </ul>
  </nav>
</template>
