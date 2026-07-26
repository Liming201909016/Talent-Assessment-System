const state = {
  singlePdfFinished: false
}
const mutations = {
  setSinglePdfFinished(state, payload){
    state.singlePdfFinished = payload;
  }
}

//
// const state = {
//   pdfStatus: false,
// };
//
// const mutations = {
//   SET_PDF_STATUS(state, pdfStatus) {
//     state.pdfStatus = pdfStatus;
//     console.log("SET_WS", state.ws);
//   },
// };
//
// const actions = {
//
//   editPdfStatus({commit, state}, pdfStatus) {
//     commit('SET_PDF_STATUS', pdfStatus);
//   },
// };
//
export default {
  state,
  mutations,
  // actions,
};
