// User web product shell.
// Connects to the local gateway when it is running; falls back to static
// mock data when it is not, so the shell stays reviewable on its own.

const API_BASE = window.API_BASE || "http://localhost:8080/api/v1";
const DEMO_USER_ID = "1"; // development auth placeholder (X-USER-ID header)

const mockAsks = [
  ["50,020.00", "0.42000000", "21,008.40"],
  ["50,010.00", "0.18000000", "9,001.80"],
  ["50,000.00", "0.10000000", "5,000.00"],
];

const mockBids = [
  ["49,990.00", "0.22000000", "10,997.80"],
  ["49,980.00", "0.36000000", "17,992.80"],
  ["49,970.00", "0.14000000", "6,995.80"],
];

let backendOnline = false;

function renderRows(target, rows, className) {
  document.getElementById(target).innerHTML = rows
    .map((row) => `<tr class="${className}"><td>${row[0]}</td><td>${row[1]}</td><td>${row[2]}</td></tr>`)
    .join("");
}

function formatNumber(value) {
  return Number(value).toLocaleString("en-US", { minimumFractionDigits: 2 });
}

function toRow(level) {
  const total = Number(level.price) * Number(level.quantity);
  return [formatNumber(level.price), level.quantity, formatNumber(total)];
}

async function api(path, options = {}) {
  const response = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      "X-USER-ID": DEMO_USER_ID,
      ...(options.headers || {}),
    },
  });
  const body = await response.json();
  if (!response.ok) {
    throw new Error(body.message || body.code || `HTTP ${response.status}`);
  }
  return body;
}

async function refreshOrderbook() {
  const book = await api("/orderbook/BTC-USDT?limit=10");
  const asks = book.asks.map(toRow).reverse();
  const bids = book.bids.map(toRow);
  renderRows("asks", asks.length ? asks : [["-", "-", "-"]], "ask");
  renderRows("bids", bids.length ? bids : [["-", "-", "-"]], "bid");
}

async function refreshLastPrice() {
  const data = await api("/trades/BTC-USDT?limit=1");
  if (data.trades.length > 0) {
    document.getElementById("last-price").textContent = formatNumber(data.trades[data.trades.length - 1].price);
  }
}

async function refreshBalances() {
  const data = await api("/account/balances");
  for (const balance of data.balances) {
    const cell = document.getElementById(`balance-${balance.asset.toLowerCase()}`);
    if (cell) {
      cell.textContent = balance.available;
    }
  }
}

async function refreshAll() {
  try {
    await Promise.all([refreshOrderbook(), refreshLastPrice(), refreshBalances()]);
    if (!backendOnline) {
      backendOnline = true;
      document.getElementById("order-result").textContent = "Connected to local gateway.";
    }
  } catch (_error) {
    if (backendOnline || !document.getElementById("asks").innerHTML) {
      backendOnline = false;
      renderRows("asks", mockAsks, "ask");
      renderRows("bids", mockBids, "bid");
      document.getElementById("order-result").textContent =
        "Gateway offline - showing mock data. Run: go run ./services/gateway/cmd/gateway";
    }
  }
}

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

document.getElementById("submit-order").addEventListener("click", async () => {
  const price = document.getElementById("price").value;
  const quantity = document.getElementById("quantity").value;
  const resultEl = document.getElementById("order-result");

  if (!backendOnline) {
    resultEl.textContent = `${side} ${quantity} BTC-USDT @ ${price} accepted by mock compliance gate.`;
    return;
  }

  try {
    const result = await api("/orders", {
      method: "POST",
      body: JSON.stringify({
        client_order_id: `web-${Date.now()}`,
        symbol: "BTC-USDT",
        side,
        type: "LIMIT",
        time_in_force: "GTC",
        price,
        quantity,
      }),
    });
    resultEl.textContent = `Order ${result.order.order_id} ${result.order.status} (${result.trades} trade(s)).`;
    await refreshAll();
  } catch (error) {
    resultEl.textContent = `Rejected: ${error.message}`;
  }
});

document.getElementById("withdraw").addEventListener("click", async () => {
  const resultEl = document.getElementById("wallet-result");
  if (!backendOnline) {
    resultEl.textContent = "Mock withdrawal queued: whitelist, KYC L2, AML Low, and admin approval required.";
    return;
  }
  try {
    await api("/wallet/withdraw", {
      method: "POST",
      body: JSON.stringify({
        id: Date.now(),
        asset: "USDT",
        network: "TESTNET",
        address: "demo-address",
        amount: "10.00",
      }),
    });
    resultEl.textContent = "Withdrawal accepted for review.";
  } catch (error) {
    resultEl.textContent = `Withdrawal blocked: ${error.message}`;
  }
});

refreshAll();
setInterval(refreshAll, 2000);
