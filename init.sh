# if ! grep -q "^GMAIL_API_TOKEN=" .env ; then
# 	docker compose run -it --build oauth_helper

# 	while [ ! -f ./compose_conf/oauth_helper/token.txt ]; do
# 		sleep 2
# 	done

# 	TOKEN=$(cat "./compose_conf/oauth_helper/token.txt")

# 	echo $'\n' >> .env
# 	echo "GMAIL_API_TOKEN=\"$TOKEN\"" >> .env
# 	rm ./compose_conf/oauth_helper/token.txt

# 	echo "token created"
# else 
# 	echo "token exists"
# fi

if [ -z "$( ls -A '/path/to/dir' )" ]; then
	docker compose run -it --build certbot
fi


while getopts f OPT 
do
	case "$OPT" in
		f) rm -rf ./compose_conf && echo "volumes removed"
	esac
done

docker compose up --build