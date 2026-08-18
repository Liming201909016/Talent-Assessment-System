<template>
  <div id="app">
    <router-view />
  </div>
</template>

<script>
export default  {
  name:  'App',
    metaInfo() {
    const hideSystemTitle = Boolean(this.$route && this.$route.meta && this.$route.meta.hideSystemTitle)
        return {
      title: hideSystemTitle ? '' : this.$store.state.settings.dynamicTitle && this.$store.state.settings.title,
            titleTemplate: title => {
        return hideSystemTitle ? '' : title ? `${title} - ${process.env.VUE_APP_TITLE}` : process.env.VUE_APP_TITLE
            }
        }
    },
}
</script>
