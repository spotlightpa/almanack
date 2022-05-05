<script>
export default {
  props: { params: Object, fileProps: Object },
};
</script>

<template>
  <div>
    <SiteParamsBanner :file-props="fileProps" :params="params" />
    <SiteParamsTopper :file-props="fileProps" :params="params" />
    <SiteParamsTop :file-props="fileProps" :params="params" />
    <SiteParamsSticky :file-props="fileProps" :params="params" />
    <SiteParamsSidebar :file-props="fileProps" :params="params" />
    <SiteParamsRiver :file-props="fileProps" :params="params" />
    <SiteParamsFooter :file-props="fileProps" :params="params" />
    <SiteParamsSupport :file-props="fileProps" :params="params" />
  </div>
</template>
