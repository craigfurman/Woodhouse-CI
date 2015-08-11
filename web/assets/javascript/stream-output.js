$(document).ready(function() {
	if (job.finished) {
		return;
	}

	var outputEvents = new EventSource('/jobs/' + job.jobId + '/builds/' + job.buildNumber + '/output?offset=' + job.bytesAleadyReceived);
	outputEvents.addEventListener("output", function(e) {
		$('#jobOutput').append(e.data);
	});

	outputEvents.addEventListener("end", function(e) {
		console.log("caught end event");
		outputEvents.close();
	});
});
