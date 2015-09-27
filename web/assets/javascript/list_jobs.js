$(document).ready(function() {
    var events = new EventSource('/jobs/status');
    events.addEventListener('jobs', function(e) {
        var statuses = JSON.parse(e.data);
        $.each(statuses, function(key, value) {
            $('#' + key).attr('class', value);
        });
    });
});
