const ui = {
    network: document.getElementById("network"),
    device: document.getElementById("device"),
    minutes: document.getElementById("minutes"),
    minutesLabel: document.getElementById("minutes-label"),
    windowPresets: document.getElementById("window-presets"),
    refreshButton: document.getElementById("refresh"),
    feedback: document.getElementById("feedback"),
    latestLatency: document.getElementById("latest-latency"),
    avgLatency: document.getElementById("avg-latency"),
    avgPacketLoss: document.getElementById("avg-packet-loss"),
    availability: document.getElementById("availability"),
    lastUpdated: document.getElementById("last-updated"),
    chartContainer: document.getElementById("chart"),
};

const AUTO_REFRESH_MS = 15000;

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

function formatLabel(isoString) {
    const date = new Date(isoString);
    if (Number.isNaN(date.getTime())) {
        return isoString;
    }
    return date.toLocaleString(undefined, {
        day: "2-digit",
        month: "short",
        hour: "2-digit",
        minute: "2-digit",
        second: "2-digit",
    });
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
            textStyle: {color: "#93a1a1"},
        },
        grid: {
            left: 44,
            right: 44,
            top: 54,
            bottom: 30,
        },
        xAxis: {
            type: "category",
            data: (series.labels || []).map(formatLabel),
            axisLabel: {color: "#657b83"},
        },
        yAxis: [
            {
                type: "value",
                name: "Latency (ms)",
                axisLabel: {color: "#2aa198"},
                splitLine: {lineStyle: {color: "rgba(131, 148, 150, 0.14)"}},
            },
            {
                type: "value",
                min: 0,
                max: 100,
                name: "Packet Loss (%)",
                axisLabel: {color: "#b58900"},
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
                connectNulls: false,
                lineStyle: {width: 2.5, color: "#2aa198"},
                areaStyle: {color: "rgba(42, 161, 152, 0.16)"},
                data: series.latency_series || [],
            },
            {
                name: "Packet Loss (%)",
                type: "line",
                yAxisIndex: 1,
                smooth: 0.22,
                showSymbol: false,
                lineStyle: {width: 2.2, color: "#b58900"},
                areaStyle: {color: "rgba(181, 137, 0, 0.14)"},
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

function setActivePreset(minutes) {
    const buttons = ui.windowPresets.querySelectorAll("button");
    for (const button of buttons) {
        button.classList.toggle("active", Number(button.dataset.minutes) === Number(minutes));
    }
}

function setWindowMinutes(minutes) {
    ui.minutes.value = minutes;
    ui.minutesLabel.textContent = `${minutes} minute${Number(minutes) === 1 ? "" : "s"}`;
    setActivePreset(minutes);
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
    setActivePreset(-1);
    ui.minutesLabel.textContent = `${ui.minutes.value} minute${Number(ui.minutes.value) === 1 ? "" : "s"}`;
});

ui.minutes.addEventListener("change", debounce(async () => {
    try {
        await loadSeries();
    } catch (error) {
        setFeedback(error.message, true);
    }
}, 250));

ui.windowPresets.addEventListener("click", async (event) => {
    const button = event.target.closest("button[data-minutes]");
    if (!button) {
        return;
    }
    setWindowMinutes(button.dataset.minutes);
    try {
        await loadSeries();
    } catch (error) {
        setFeedback(error.message, true);
    }
});

ui.refreshButton.addEventListener("click", async () => {
    try {
        await loadSeries();
    } catch (error) {
        setFeedback(error.message, true);
    }
});

window.addEventListener("resize", () => chart.resize());

setInterval(() => {
    if (document.hidden) {
        return;
    }
    loadSeries().catch((error) => setFeedback(error.message, true));
}, AUTO_REFRESH_MS);

bootstrap();
