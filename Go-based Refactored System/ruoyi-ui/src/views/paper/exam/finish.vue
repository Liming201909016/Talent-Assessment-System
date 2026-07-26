<template>
  <div class="text-container">
    <p>感谢您的参与!</p>
  </div>
</template>

<script>


export default {
  name: 'Finish',

  data() {
    return {

    }
  },

  mounted() {

    history.pushState(null, null, document.URL);

    window.addEventListener("popstate",function(e) {
      history.pushState(null, null, document.URL);
    }, false);
  },

  methods: {
    //body {
    //  display: flex;
    //  justify-content: center;
    //  align-items: center;
    //  height: 100vh;
    //}
    //
    //.text-container {
    //  display: flex;
    //  justify-content: center;
    //  align-items: center;
    //}
    //
    //p {
    //  text-align: center;
    //}

  }
}
</script>

<style scoped>

  .text-container{
    position: absolute;
    top: 50%;
    left: 50%;
    -webkit-transform: translate(-50%, -50%);
    -moz-transform: translate(-50%, -50%);
    -ms-transform: translate(-50%, -50%);
    -o-transform: translate(-50%, -50%);
    transform: translate(-50%, -50%);
  }

  p {
    font-size: 50px;
    color: #36e010;
  }

</style>


