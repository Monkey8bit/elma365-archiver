while getopts f OPT 
do
	case "$OPT" in
		f) rm -rf ./compose_conf && echo "volumes removed"
	esac
done	

docker compose up --build