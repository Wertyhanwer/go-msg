import dataclasses
import subprocess
from typing import List


@dataclasses.dataclass
class _CommandInfo:
	message: str
	run: List[str] | None = None


_MAP_CMD_TO_INFO: dict[str, _CommandInfo] = {
	str(index): cmd_info for  index, cmd_info in enumerate(
		(
			_CommandInfo(
				"Exit 'setup.py'.", None
			),
			_CommandInfo(
				"Reassembles the images and launches the containers (docker-compose up --build).",
				["docker-compose", "up", "--build"]
			),
			_CommandInfo(
				"Stops and deletes all containers (docker-compose down)..",
				["docker-compose", "down"]
			),
			_CommandInfo(
				"Shows a list of running containers and their statuses (docker-compose ps).",
				["docker-compose", "ps"]
			),
		),
		start=0
	)
}


def main():
	print("Welcome to the project settings")

	while True:
		print("\nMenu:")
		for key in sorted(_MAP_CMD_TO_INFO.keys()):
			print(f"{key}) {_MAP_CMD_TO_INFO[key].message}")

		user_input = input("\nInput number: ")
		print("")

		if user_input not in _MAP_CMD_TO_INFO:
			print("You entered an invalid command. Please try again.")
			continue

		cmd_info = _MAP_CMD_TO_INFO[user_input]
		print(cmd_info.message)

		if cmd_info.run is not None:
			subprocess.run(cmd_info.run)

		break


if __name__ == "__main__":
	main()
