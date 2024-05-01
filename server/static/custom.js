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

// no it should be detect single lady function !
function detectSingleDevice() {
    let optionsCount = $('#device option').length;
    console.log(optionsCount);
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
        let data = {"name": this.value};
        console.log(data)
        // commonAjax(data, '/device/', function (data, error){
        //     if (error) {
        //         console.error("Network call error: ", error);
        //     } else {
        //         commonDeviceSelector(data);
        //     }
        // });
    });
}

$(document).ready(function(){
    networkSelector();
    detectSingleNetwork();
    deviceSelector();
});
