# Flat File "Secure" Database

Sometimes you have a database with information of your users (**not sensible one!!!**) or something else. If this database and your application have a lot of requests you could have troubles.

In this cases, you could go to AWS, GCE or any other cloud, or you could create a system of hashed files from a CSV, from a combination of information that only your customer know (id + some password, e.g.) or known fixed information (selectable info to avoid free text in input boxes).

# How does it work

We need to setup some properties:

		filename: ./data/yourdatabase.csv
		delimiter: ; --> if the value is comma ',' better wrap it with quotes
		outputdir: ./html/output
		deepdirs: 3 --> this will create a structure of directories of N deep
		max_threads: 150 --> max workers to deal with file creation
		pbkdf2_iterations: 1000 --> the derived key is using pbkdf2 algorith, then, we need some iterations 
		pbkdf2_keylength: 32 
		hash_dir: false --> if you want to apply a pbkdf2 derivation to the folder tree created from the first N chars of filenames

		colums_hash: 1,2 --> colums from the csv to perform the hashing
		columns_content: 3,4 --> columns from the csv to put in the hashed file. It could be a negative integer, then, all the columns from the positive integer will be added to the content of the hashed file. E.g. -3 will add all columns from the 3rd to the end.


# Build or run

Build or run the code to try it:

		$ go run main.go

You can pass a config file adding a flag

		$ go run main.go -config=youconfigfile.yaml

This will generate our file structure and you will be able to call this data from your html client doing the same hashing process that we do with the process.

# Test it

1. Run main program
		
		$ go run main.go

1. When finish, run http-server

		$ go run http-server.go

1. Go to [http://localhost](http://localhost)

	- There are two samples:
	
		- Default config.yaml: the sample is built from a csv with spanish localities. If you put "08" in "col1" and "135" in "col2" it will work. 
		- dvds.yaml: the sample is built from a csv from http://www.hometheaterinfo.com/dvdlist.htm. The keys are "id" and "year" of the movie. Try "032429216816" and "2013" to get "World War Z"

	Imagine only you know that data. 

# HTML

It is based entirely on http://anandam.name/pbkdf2/.

You only have to setup the "config" object in the same way you did with config-yaml:

	<script src="https://code.jquery.com/jquery-2.1.4.min.js"></script>
	<script src="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.5/js/bootstrap.min.js"></script>
	<script src="/js/sha1.js"></script>
	<script src="/js/pbkdf2.js"></script>
	<script>

		var config = {
			deepdirs : 1,
			pbkdf2_iterations : 1,
			pbkdf2_keylength : 32,
			hash_dir : false,
			datadir : "/output/",
			fields_to_query : ["Col 1","Col 2"],
			statusid : "status"	
		}

		var fieldTemplate = [
			'<div class="form-group"',
			'	<label for="{{id}}" id="l_{{id}}"><span>{{id}}</span>:</label>',
			'	<input type="text" class="form-control" id="{{id}}" autocomplete="off" />',
			'</div>'
		].join("\n");

		function request(path){
			$.ajax({
				url : path,
		       	contentType: "text/plain;charset=iso-8859-15",

			})
			.done(function( data ) {
			  data = data.split(",")
			  $("#"+config.statusid+" .modal-body strong").html(data.join(" "));
			})
			.error(function(){
			  $("#"+config.statusid+" .modal-body strong").html("No data");
			});
		}

		$(document).ready(function(){
			var aux="";
			for(var i=0;i<config.fields_to_query.length;i++){
				aux = config.fields_to_query[i].replace(" ","_")
				$(fieldTemplate.replace(/{{id}}/g,aux)).insertBefore($("#getInfo button"));
				aux=$("#l_"+aux+ " span").html();
				$("#l_"+aux + " span").html(aux.replace("_", " "));
			}

			$('.modal').on('hidden.bs.modal', function () {
				$("#"+config.statusid+" .modal-body strong").html('<span class="glyphicon glyphicon-refresh spinning"></span>');
			});

			$("#getInfo").submit(function(){

   			    $("#"+config.statusid).modal()
				
				var hash = "";

				for(var i=0;i<config.fields_to_query.length;i++){
					hash = hash + $("#"+(config.fields_to_query[i].replace(" ","_"))).val();
				}

				var filehash = new PBKDF2(hash+hash, hash+hash+hash, config.pbkdf2_iterations, config.pbkdf2_keylength);
								
				var filehash_callback = function(key) {
	    

					if(config.hash_dir){
						var folder = new PBKDF2(key.slice(0,config.deepdirs), hash+hash+hash, config.pbkdf2_iterations, config.deepdirs);
						folder.deriveKey(function(percent_done){}, function(dkfolder){
							path = config.datadir+dkfolder.slice(0,config.deepdirs).split("").join("/")+"/"+key;
							request(path);
						});
					}else{
						request(config.datadir+key.slice(0,config.deepdirs).split("").join("/")+"/"+key)
					}

				};


				filehash.deriveKey(function(percent_done){}, filehash_callback);
				return false;
			});

		});
	</script>

# Benchmark

* VirtualBox VM Ubuntu 15.04, 1GB RAM, 1vCPU

* Processing:
	- /data/municipios.csv (8.119 rows)
		- 1 deep dir
		- only 1 pbkdf2 iteration
		- 300 threads

		It takes 2'' aprox.

	- /data/dvd_csv.txt (282.404 rows)
		- 3 deep dir
		- only 1 pbkdf2 iteration
		- 500 threads

		It takes 3' 40'' aprox.
