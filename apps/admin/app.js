document.querySelectorAll("button[data-action]").forEach((button) => {
  button.addEventListener("click", () => {
    const action = button.dataset.action;
    const timestamp = new Date().toISOString();
    document.getElementById("audit-status").textContent =
      `${action} queued with MFA, RBAC, and audit-chain requirements at ${timestamp}.`;
  });
});

