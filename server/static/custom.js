function commonAjax(data, path, callback) {
    $.ajax({
        type: "POST",
        url: path,
        data: JSON.stringify(data),
        contentType: "application/json",
        success: function (data) {
            callback(data);
        },
        error: function (error) {
            callback(null, error);
        }
    })
}

function commonDeviceSelector(data) {
    $("#device").empty(); // delete any previous options
    if (Object.keys(data).length > 0) { // check network has any device
        for (let key in data) {
            $('<option/>', { value : data[key]["name"] }).text(data[key]["name"]).appendTo('#device');
        }
    } else {
        $('<option/>', { value : "" }).text("This network doesn't have any associated devices.").appendTo('#device');
    }
}


// handles network selection and device selector generation based on the input value.
function networkSelector() {
    $('#network').on('change', function() {
        let data = {"name": this.value};
        commonAjax(data, '/device/', function (data, error){
            if (error) {
                console.error("Network call error: ", error);
            } else {
                commonDeviceSelector(data);
            }
        });
    });
}

function generateChart(data) {
    $('#loader').remove();
    let chart = echarts.init(document.getElementById('chart'), null, {
        renderer: 'svg'
    });
    console.log(data.series);
    let option = {
        title: {
            text: 'Temperature Change in the Coming Week'
        },
        tooltip: {
            trigger: 'axis'
        },
        legend: {},
        toolbox: {
            show: true,
            feature: {
                dataZoom: {
                    yAxisIndex: 'none'
                },
                dataView: { readOnly: false },
                magicType: { type: ['line', 'bar'] },
                restore: {},
                saveAsImage: {}
            }
        },
        xAxis: {
            type: 'category',
            data: data.labels
        },
        yAxis: {
            type: 'value'
        },
        series: [{
            data: data.series,
            type: 'line',
            smooth: true
        }]
    };
    chart.setOption(option);
}

function getChartData() {
    let network, device;
    let networkSelector = $('#network');
    let deviceSelector = $('#device');

    network = networkSelector.val();
    device = deviceSelector.val();

    if(!network) {
        network = networkSelector.find("option:first-child").val();
    }
    if(!device) {
        network = deviceSelector.find("option:first-child").val();
    }

    let data = {"network_name": network, "device_name": device};
    commonAjax(data, '/chart/', function (data, error){
        if (error) {
            console.error("get chart data error: ", error);
        } else {
            console.log(data);
            generateChart(data);
        }
    });
}

// no it should be detect single lady function !
function detectSingleDevice() {
    let optionsCount = $('#device option').length;
    if (optionsCount === 1 ) {
        getChartData();
    }
}

// if the user has only one network, they don't need to select a single network
function detectSingleNetwork() {
    let optionsCount = $('#network option').length;
    if (optionsCount === 1 ) {
        let data = {"name": $('#network').find("option:first-child").val()};
        commonAjax(data, '/device/', function (data, error){
            if (error) {
                console.error("Single network call error: ", error);
            } else {
                commonDeviceSelector(data);
                detectSingleDevice()
            }
        });
    }
}

function deviceSelector() {
    $('#device').on('change', function() {
        getChartData();
    });
}

$(document).ready(function(){
    networkSelector();
    detectSingleNetwork();
    deviceSelector();
});
