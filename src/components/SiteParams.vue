<script>
export default {
  props: { params: Object, fileProps: Object },
  setup(props) {
    return {
      update([name, value]) {
        props.params.data[name] = value;
      },
    };
  },
};
</script>

<template>
  <div>
    <SiteParamsBanner :file-props="fileProps" :params="params" @data="update" />
    <SiteParamsTopper :file-props="fileProps" :params="params" @data="update" />
    <SiteParamsTop :file-props="fileProps" :params="params" @data="update" />
    <SiteParamsSticky :file-props="fileProps" :params="params" @data="update" />
    <SiteParamsSidebar
      :file-props="fileProps"
      :params="params"
      @data="update"
    />
    <SiteParamsRiver :file-props="fileProps" :params="params" @data="update" />
    <SiteParamsFooter :file-props="fileProps" :params="params" @data="update" />
    <SiteParamsSupport
      :file-props="fileProps"
      :params="params"
      @data="update"
    />
  </div>
</template>
