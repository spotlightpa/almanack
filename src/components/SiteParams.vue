<script setup>
import { ref } from "vue";

const props = defineProps({ params: Object, fileProps: Object });

const adBreaker = ref();
const adFeatured = ref();
const adHeadwater = ref();
const adRail = ref();
const banner = ref();
const header = ref();
const featuredArticle = ref();
const footer = ref();
const newsletter = ref();
const river = ref();
const sidebar = ref();
const sticky = ref();
const support = ref();
const takeover = ref();

function saveParams() {
  return {
    schedule_for: props.params.scheduleFor,
    data: {
      ...props.params.data,
      ...adBreaker.value.saveData(),
      ...adFeatured.value.saveData(),
      ...adHeadwater.value.saveData(),
      ...adRail.value.saveData(),
      ...banner.value.saveData(),
      ...header.value.saveData(),
      ...featuredArticle.value.saveData(),
      ...footer.value.saveData(),
      ...newsletter.value.saveData(),
      ...river.value.saveData(),
      ...sidebar.value.saveData(),
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
    ref="adFeatured"
    :file-props="fileProps"
    :params="params"
  ></SiteParamsFeaturedHomepage>
  <SiteParamsBreaker
    ref="adBreaker"
    :file-props="fileProps"
    :params="params"
  ></SiteParamsBreaker>
  <SiteParamsHeadwater
    ref="adHeadwater"
    :file-props="fileProps"
    :params="params"
  ></SiteParamsHeadwater>
  <SiteParamsRail
    ref="adRail"
    :file-props="fileProps"
    :params="params"
  ></SiteParamsRail>
  <SiteParamsRiver
    ref="river"
    :file-props="fileProps"
    :params="params"
  ></SiteParamsRiver>

  <h2 class="mt-5 title is-5">Article page ad slots</h2>
  <SiteParamsSidebar
    ref="sidebar"
    :file-props="fileProps"
    :params="params"
  ></SiteParamsSidebar>
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
