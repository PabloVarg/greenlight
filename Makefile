run:
	@ docker compose up --detach
	@ air
	@ killall -9 air
