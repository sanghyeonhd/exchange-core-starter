const asks = [
  ["50,020.00", "0.42000000", "21,008.40"],
  ["50,010.00", "0.18000000", "9,001.80"],
  ["50,000.00", "0.10000000", "5,000.00"],
];

const bids = [
  ["49,990.00", "0.22000000", "10,997.80"],
  ["49,980.00", "0.36000000", "17,992.80"],
  ["49,970.00", "0.14000000", "6,995.80"],
];

function renderRows(target, rows, className) {
  document.getElementById(target).innerHTML = rows
    .map((row) => `<tr class="${className}"><td>${row[0]}</td><td>${row[1]}</td><td>${row[2]}</td></tr>`)
    .join("");
}

renderRows("asks", asks, "ask");
renderRows("bids", bids, "bid");

let side = "BUY";
document.getElementById("buy-tab").addEventListener("click", () => {
  side = "BUY";
  document.getElementById("buy-tab").classList.add("active");
  document.getElementById("sell-tab").classList.remove("active");
});

document.getElementById("sell-tab").addEventListener("click", () => {
  side = "SELL";
  document.getElementById("sell-tab").classList.add("active");
  document.getElementById("buy-tab").classList.remove("active");
});

document.getElementById("submit-order").addEventListener("click", () => {
  const price = document.getElementById("price").value;
  const quantity = document.getElementById("quantity").value;
  document.getElementById("order-result").textContent =
    `${side} ${quantity} BTC-USDT @ ${price} accepted by mock compliance gate.`;
});

document.getElementById("withdraw").addEventListener("click", () => {
  document.getElementById("wallet-result").textContent =
    "Mock withdrawal queued: whitelist, KYC L2, AML Low, and admin approval required.";
});

