document.addEventListener("DOMContentLoaded", () => {
  const codeBlocks = document.querySelectorAll("pre");

  codeBlocks.forEach((pre) => {
    const container = document.createElement("div");
    container.className = "copy-code-container";

    const button = document.createElement("button");
    button.className = "copy-code-btn";
    button.textContent = "Copy";

    pre.parentNode.insertBefore(container, pre);
    container.appendChild(pre);
    container.appendChild(button);

    button.addEventListener("click", () => {
      const code = pre.querySelector("code");
      if (!code) return;

      const text = code.innerText;

      navigator.clipboard
        .writeText(text)
        .then(() => {
          button.textContent = "Copied!";
          setTimeout(() => {
            button.textContent = "Copy";
          }, 2000);
        })
        .catch((err) => {
          console.error("Failed to copy: ", err);
          button.textContent = "Error";
        });
    });
  });
});
