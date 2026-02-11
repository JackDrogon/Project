.PHONY: build clean run

build:
	@mkdir -p build && cd build && cmake .. && $(MAKE)

clean:
	@rm -rf build

run: build
	@./build/{{.ProjectName}}
