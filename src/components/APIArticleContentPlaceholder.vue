<script>
export default {
  props: {
    block: Object
  }
};
</script>

<template>
  <h2 style="color: red;">
    Embed <a :href="`#inline-embed-${block.n}`"></a>#{{ block.n }}
  </h2>
</template>
