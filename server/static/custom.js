const ui = {
    network: document.getElementById("network"),
    device: document.getElementById("device"),
    minutes: document.getElementById("minutes"),
    minutesLabel: document.getElementById("minutes-label"),
    refreshButton: document.getElementById("refresh"),
    feedback: document.getElementById("feedback"),
    latestLatency: document.getElementById("latest-latency"),
    avgLatency: document.getElementById("avg-latency"),
    avgPacketLoss: document.getElementById("avg-packet-loss"),
    availability: document.getElementById("availability"),
    lastUpdated: document.getElementById("last-updated"),
    chartContainer: document.getElementById("chart"),
};

const chart = echarts.init(ui.chartContainer, null, {renderer: "svg"});

function setFeedback(message, isError = false) {
    ui.feedback.classList.toggle("error", isError);
    ui.feedback.textContent = message;
}

function toFixed(value, digits = 2) {
    if (typeof value !== "number" || Number.isNaN(value)) {
        return "-";
    }
    return value.toFixed(digits);
}

async function fetchJSON(url) {
    const response = await fetch(url);
    if (!response.ok) {
        throw new Error(`request failed: ${response.status}`);
    }
    return response.json();
}

function setOptions(select, items) {
    select.innerHTML = "";
    for (const item of items) {
        const option = document.createElement("option");
        option.value = item;
        option.textContent = item;
        select.appendChild(option);
    }
}

async function loadNetworks() {
    const data = await fetchJSON("/api/networks");
    const items = data.items || [];
    if (items.length === 0) {
        throw new Error("No networks found. Add one with `./main network add`.");
    }
    setOptions(ui.network, items);
}

async function loadDevices(networkName) {
    const data = await fetchJSON(`/api/networks/${encodeURIComponent(networkName)}/devices`);
    const items = data.items || [];
    if (items.length === 0) {
        throw new Error("Selected network has no devices.");
    }
    setOptions(ui.device, items);
}

function renderChart(series) {
    chart.setOption({
        animationDuration: 500,
        tooltip: {
            trigger: "axis",
        },
        legend: {
            data: ["Latency (ms)", "Packet Loss (%)"],
            textStyle: {color: "#eff9ff"},
        },
        grid: {
            left: 44,
            right: 44,
            top: 54,
            bottom: 30,
        },
        xAxis: {
            type: "category",
            data: series.labels || [],
            axisLabel: {color: "#b5d4df"},
        },
        yAxis: [
            {
                type: "value",
                name: "Latency (ms)",
                axisLabel: {color: "#47f0d0"},
                splitLine: {lineStyle: {color: "rgba(255,255,255,0.08)"}},
            },
            {
                type: "value",
                min: 0,
                max: 100,
                name: "Packet Loss (%)",
                axisLabel: {color: "#ffca6f"},
                splitLine: {show: false},
            }
        ],
        series: [
            {
                name: "Latency (ms)",
                type: "line",
                yAxisIndex: 0,
                smooth: 0.28,
                showSymbol: false,
                lineStyle: {width: 2.5, color: "#45f0c2"},
                areaStyle: {color: "rgba(69, 240, 194, 0.16)"},
                data: series.latency_series || [],
            },
            {
                name: "Packet Loss (%)",
                type: "line",
                yAxisIndex: 1,
                smooth: 0.22,
                showSymbol: false,
                lineStyle: {width: 2.2, color: "#ffd166"},
                areaStyle: {color: "rgba(255, 209, 102, 0.14)"},
                data: series.packet_loss_series || [],
            }
        ]
    });
}

function renderSummary(summary) {
    ui.latestLatency.textContent = `${toFixed(summary.latest_latency_ms)} ms`;
    ui.avgLatency.textContent = `${toFixed(summary.average_latency_ms)} ms`;
    ui.avgPacketLoss.textContent = `${toFixed(summary.average_packet_loss_percent)} %`;
    ui.availability.textContent = `${toFixed(summary.availability_percent)} %`;

    if (summary.last_updated) {
        const date = new Date(summary.last_updated);
        ui.lastUpdated.textContent = `Last updated ${date.toLocaleString()}`;
    } else {
        ui.lastUpdated.textContent = "No recent samples";
    }
}

async function loadSeries() {
    const network = ui.network.value;
    const device = ui.device.value;
    const minutes = ui.minutes.value;
    if (!network || !device) {
        return;
    }

    setFeedback("Loading telemetry...");
    const params = new URLSearchParams({network, device, minutes});
    const data = await fetchJSON(`/api/series?${params.toString()}`);
    renderChart(data.series);
    renderSummary(data.summary);
    setFeedback(`Showing ${data.summary.samples} samples`);
}

function debounce(fn, wait) {
    let timeout = null;
    return (...args) => {
        clearTimeout(timeout);
        timeout = setTimeout(() => fn(...args), wait);
    };
}

async function bootstrap() {
    try {
        await loadNetworks();
        await loadDevices(ui.network.value);
        await loadSeries();
    } catch (error) {
        setFeedback(error.message, true);
    }
}

ui.network.addEventListener("change", async () => {
    try {
        await loadDevices(ui.network.value);
        await loadSeries();
    } catch (error) {
        setFeedback(error.message, true);
    }
});

ui.device.addEventListener("change", async () => {
    try {
        await loadSeries();
    } catch (error) {
        setFeedback(error.message, true);
    }
});

ui.minutes.addEventListener("input", () => {
    const value = Number(ui.minutes.value);
    ui.minutesLabel.textContent = `${value} minute${value === 1 ? "" : "s"}`;
});

ui.minutes.addEventListener("change", debounce(async () => {
    try {
        await loadSeries();
    } catch (error) {
        setFeedback(error.message, true);
    }
}, 250));

ui.refreshButton.addEventListener("click", async () => {
    try {
        await loadSeries();
    } catch (error) {
        setFeedback(error.message, true);
    }
});

window.addEventListener("resize", () => chart.resize());
bootstrap();
