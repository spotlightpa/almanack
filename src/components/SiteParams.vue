<script setup>
import { ref } from "vue";

const props = defineProps({ params: Object, fileProps: Object });

const banner = ref();
const breaker = ref();
const featuredArticle = ref();
const featuredHomepage = ref();
const footer = ref();
const header = ref();
const headwater = ref();
const newsletter = ref();
const railArticle = ref();
const railHomepage = ref();
const river = ref();
const sticky = ref();
const support = ref();
const takeover = ref();

function saveParams() {
  return {
    schedule_for: props.params.scheduleFor,
    data: {
      ...props.params.data,
      ...banner.value.saveData(),
      ...breaker.value.saveData(),
      ...featuredArticle.value.saveData(),
      ...featuredHomepage.value.saveData(),
      ...footer.value.saveData(),
      ...header.value.saveData(),
      ...headwater.value.saveData(),
      ...newsletter.value.saveData(),
      ...railArticle.value.saveData(),
      ...railHomepage.value.saveData(),
      ...river.value.saveData(),
      ...sticky.value.saveData(),
      ...support.value.saveData(),
      ...takeover.value.saveData(),
    },
  };
}

defineExpose({
  saveParams,
});
</script>

<template>
  <h2 class="title is-5">General ad slots</h2>
  <SiteParamsBanner
    ref="banner"
    :file-props="fileProps"
    :params="params"
  ></SiteParamsBanner>
  <SiteParamsHeader
    ref="header"
    :file-props="fileProps"
    :params="params"
  ></SiteParamsHeader>
  <SiteParamsFooter
    ref="footer"
    :file-props="fileProps"
    :params="params"
  ></SiteParamsFooter>

  <h2 class="mt-5 title is-5">Homepage ad slots</h2>
  <SiteParamsFeaturedHomepage
    ref="featuredHomepage"
    :file-props="fileProps"
    :params="params"
  ></SiteParamsFeaturedHomepage>
  <SiteParamsBreaker
    ref="breaker"
    :file-props="fileProps"
    :params="params"
  ></SiteParamsBreaker>
  <SiteParamsHeadwater
    ref="headwater"
    :file-props="fileProps"
    :params="params"
  ></SiteParamsHeadwater>
  <SiteParamsRailHomepage
    ref="railHomepage"
    :file-props="fileProps"
    :params="params"
  ></SiteParamsRailHomepage>
  <SiteParamsRiver
    ref="river"
    :file-props="fileProps"
    :params="params"
  ></SiteParamsRiver>

  <h2 class="mt-5 title is-5">Article page ad slots</h2>
  <SiteParamsRailArticle
    ref="railArticle"
    :file-props="fileProps"
    :params="params"
  ></SiteParamsRailArticle>
  <SiteParamsSupport
    ref="support"
    :file-props="fileProps"
    :params="params"
  ></SiteParamsSupport>
  <SiteParamsFeaturedArticle
    ref="featuredArticle"
    :file-props="fileProps"
    :params="params"
  ></SiteParamsFeaturedArticle>

  <h2 class="mt-5 title is-5">Sticky and popup ad slots</h2>
  <SiteParamsSticky
    ref="sticky"
    :file-props="fileProps"
    :params="params"
  ></SiteParamsSticky>
  <SiteParamsNewsletter
    ref="newsletter"
    :file-props="fileProps"
    :params="params"
  ></SiteParamsNewsletter>
  <SiteParamsTakeover
    ref="takeover"
    :file-props="fileProps"
    :params="params"
  ></SiteParamsTakeover>
</template>

<style>
.ad-demo {
  max-width: 100%;
  width: 500px;
  margin-inline: auto;
  display: block;
}
</style>
