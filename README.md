# Flat File "Secure" Database

Sometimes you have a database with information of your users (**not sensible one!!!**) or something else. If this database and your application have a lot of requests you could have troubles.

In this cases, you could go to AWS, GCE or any other cloud, or you could create a system of hashed files from a CSV, from a combination of information that only your customer know (id + some password, e.g.) or known fixed information (selectable info to avoid free text in input boxes).

# How does it works

We need to setup some properties:

		filename: ./data/yourdatabase.csv
		delimiter: ;
		outputdir: ./html/output
		deepdirs: 3 --> this will create a structure of directories of N deep
		max_threads: 150 --> max workers to deal with file creation
		pbkdf2_iterations: 1000 --> the derived key is using pbkdf2 algorith, then, we need some iterations 
		pbkdf2_keylength: 32 
		hash_dir: false --> if you want to apply a pbkdf2 derivation to the folder tree created from N chars of filenames

		colums_hash: 1,2 --> colums from the csv to perform the hashing
		columns_content: 3,4 --> columns from the csv to put in the hashed file


# Build or run

Build or run the code to try it:

		$ go run main.go


This will generate our file structure and you will be able to call this data from your html client doing the same hashing process that we do with the process.

# HTML

It is based entirely on http://anandam.name/pbkdf2/.

You only have to setup some properties to derive the key you are searching in the same way that the Go process.

	<script src="https://code.jquery.com/jquery-2.1.4.min.js"></script>
	<script src="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.5/js/bootstrap.min.js"></script>
	<script src="/js/sha1.js"></script>
	<script src="/js/pbkdf2.js"></script>
	<script>
		var deep = 3;
		var iter = 1000;
		var length = 32;
		var datadir = "/output/";
		var number_csv_hash_fields = 2;
		var hash_dir = false;
		var statusid = "status";

		var fieldTemplate = [
			'<div class="form-group"',
			'	<label for="col{{id}}">Col {{id}}:</label>',
			'	<input type="text" class="form-control" id="col{{id}}" />',
			'</div>'
		].join("\n");

		function request(path){
			$.ajax({
				url : path,
		       	contentType: "text/plain;charset=iso-8859-15",

			})
			.done(function( data ) {
			  data = data.split(",")
			  $("#"+statusid+" .modal-body strong").html(data.join(" "));
			  $("#"+statusid).modal()
			})
			.error(function(){
			  $("#"+statusid+" .modal-body strong").html("No data");
			  $("#"+statusid).modal()
			});
		}

		$(document).ready(function(){

			for(var i=0;i<number_csv_hash_fields;i++){
				$(fieldTemplate.replace(/{{id}}/g,(i+1))).insertBefore($("#getInfo button"))
			}

			$("#getInfo").submit(function(){

				var hash = "";

				for(var i=0;i<number_csv_hash_fields;i++){
					hash = hash + $("#col"+(i+1)).val();
				}

				var mypbkdf2 = new PBKDF2(hash+hash, hash+hash+hash, iter, length);
								
				var result_callback = function(key) {
	    

					if(hash_dir){
						var folder = new PBKDF2(key.slice(0,deep), hash+hash+hash, iter, deep);
						folder.deriveKey(function(percent_done){}, function(dkfolder){
							path = datadir+dkfolder.slice(0,deep).split("").join("/")+"/"+key;
							request(path);
						});
					}else{
						request(datadir+key.slice(0,deep).split("").join("/")+"/"+key)
					}

				};


				mypbkdf2.deriveKey(function(percent_done){}, result_callback);
				return false;
			});

		});
	</script>


# Example

1. Run main program
		
		$ go run main.go

1. When finish, run http-server

		$ go run http-server.go

1. Go to http://localhost:3000

	The sample is built from a csv with spanish localities. If you put "08" in "col1" and "135" in "col2" it will work. Imagine only you know that data. 

