# For usage information, invoke this makefile without a target.
.DEFAULT_GOAL = help

all: help

# Generate the protobuf code.
gen:
	protoc --go_out=my_proto_stuff --proto_path=. --go_opt=paths=source_relative ./my_proto_stuff.proto

.PHONY: help
# Print help message.
help:
	@awk 'BEGIN{ds=1;ks=0;ms=25;c="";td=ENVIRON["TMPDIR"];if(td==""){td="/tmp"}sf=sprintf("%s/sort_%d.txt",td,int(1000000*rand()));}function pr(d,tgt,f,v){v=d[tgt];gsub(/\n/,"\n"ind pad,v);printf(f,tgt,v)}/^[[:blank:]]*#[[:blank:]]?/{tc=$$0;sub(/^[[:blank:]]*#[[:blank:]]?/,"",tc);if(c!=""){c=c"\n"}c=c tc;next}/^[[:blank:]]*[^[:blank:]:.=?%][^[:blank:]:=?%]*[[:blank:]]*:/{tgt=$$0;sub(/[[:blank:]]*:.*$$/,"",tgt);sub(/^[[:blank:]]*/,"",tgt);if(c==""){next}if(length(tgt)>ks){ks=length(tgt)}ca[tgt]=c;print tgt>sf;c="";next}//{c=""}END{close(sf);if(ks>ms){ks=ms}ind="    ";f=ind"%-"ks"s  %s\n";pad=sprintf("%"ks"s  "," ");printf("Usage: make <target>\n\n  Targets:\n");if(ds==1){cmd="sort "sf;while((cmd|getline tgt)>0){pr(ca,tgt,f)}close(cmd)}else{while((getline tgt<sf)>0){pr(ca,tgt,f)}close(sf)}system("/bin/rm "sf)}' $(firstword $(MAKEFILE_LIST))
