
// handles form selection and select generation based on the input value. The code is ugly as hell, but idc
$('#network').on('change', function() {
    let data = {"name": this.value};
    $.ajax({
        type: "POST",
        url: "/device/",
        data: JSON.stringify(data),
        contentType: "application/json",
        success: function (data) {
            $("#device").empty(); // delete any previous options
            if (Object.keys(data).length > 0) {
                for (let key in data) {
                    $('<option/>', { value : data[key]["name"] }).text(data[key]["name"]).appendTo('#device');
                }
            } else {
                $('<option/>', { value : "" }).text("This network doesn't have any associated devices.").appendTo('#device');
            }


        }
    });
});