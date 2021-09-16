<script>
export default {
  props: { links: Array },
};
</script>

<template>
  <nav class="breadcrumb has-succeeds-separator" aria-label="breadcrumbs">
    <ul>
      <li
        v-for="(link, i) of links"
        :key="i"
        :class="{ 'is-active': i === links.length - 1 }"
      >
        <router-link exact :to="link.to" v-text="link.name" />
      </li>
    </ul>
  </nav>
</template>
