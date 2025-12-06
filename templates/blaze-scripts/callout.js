// Handle collapsible callouts
document.addEventListener("DOMContentLoaded", function () {
  const callouts = document.querySelectorAll(".callout[data-callout-fold]");

  callouts.forEach((callout) => {
    const title = callout.querySelector(".callout-title");
    if (!title) return;

    title.addEventListener("click", function () {
      const currentState = callout.getAttribute("data-callout-fold");
      if (currentState === "collapsed") {
        callout.setAttribute("data-callout-fold", "expanded");
      } else {
        callout.setAttribute("data-callout-fold", "collapsed");
      }
    });
  });
});
