# Конфигурация
##############################################################################

path := .
DOCKER_COMPOSE := docker-compose
DOCKER := docker

# Определения цветов для вывода в терминал (macOS не работает)
RED    := \\033[0;31m
GREEN  := \\033[1;32m
YELLOW := \\033[1;33m
BLUE   := \\033[0;36m
RESET  := \\033[0m


# Определения окружения
##############################################################################

ifneq ($(wildcard .env),)
    ENV_FILE := .env
    ENV_SOURCE := $(YELLOW).env file$(RESET)
else
    ENV_FILE :=
    ENV_SOURCE := $(YELLOW)default values$(RESET)
endif

ifneq ($(wildcard .git/HEAD),)
    CURRENT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
else
    CURRENT_BRANCH := main
endif

ifdef PROFILE
    PROFILE := $(PROFILE)
else
    ifeq ($(CURRENT_BRANCH),main)
        PROFILE := prod
    else
        PROFILE := dev
    endif
endif

ifeq ($(PROFILE),prod)
    ENV_NAME := production
else
    ENV_NAME := development
endif

# Основные команды
##############################################################################

.PHONY: run prod dev stop clean help

.DEFAULT_GOAL := help

run: check-env  ## Launch app with auto-detected profile
	@echo -e "\n$(BLUE)Starting $(ENV_NAME) environment ($(ENV_SOURCE))...$(RESET)"
	@echo -e "$(GREEN)Active server profile:$(RESET) $(YELLOW)$(PROFILE)$(RESET)"
	@$(DOCKER_COMPOSE) $(if $(ENV_FILE),--env-file $(ENV_FILE),) --profile $(PROFILE) up --build

prod:  ## Launch production environment
	@$(MAKE) run PROFILE=prod

dev:  ## Launch development environment
	@$(MAKE) run PROFILE=dev

stop:  ## Stop all containers
	@echo -e "\n$(YELLOW)Stopping containers...$(RESET)"
	@$(DOCKER_COMPOSE) down

clean: stop  ## Remove containers and volumes
	@echo -e "\n$(RED)Removing containers and volumes...$(RESET)"
	@$(DOCKER_COMPOSE) down -v
	@echo -e "$(GREEN)Clean complete!$(RESET)"

# Верификация использования команд
##############################################################################

.PHONY: check-env check-docker

check-env: check-docker
	@echo -e "$(GREEN)Environment:$(RESET) $(YELLOW)$(ENV_NAME)$(RESET)"
	@echo -e "$(GREEN)Server profile:$(RESET) $(YELLOW)$(PROFILE)$(RESET)"
	@echo -e "$(GREEN)Configuration source:$(RESET) $(ENV_SOURCE)"
	@echo -e "$(GREEN)Branch:$(RESET) $(YELLOW)$(CURRENT_BRANCH)$(RESET)"

check-docker:
	@which $(DOCKER_COMPOSE) > /dev/null || \
		(echo -e "$(RED)Error: $(DOCKER_COMPOSE) not found!$(RESET)" && exit 1)

# Помощь по командам
##############################################################################

help:  ## Show this help message
	@echo -e "\n$(BLUE)Available targets:$(RESET)"
	@echo -e "------------------"
	@awk 'BEGIN {FS = ":.*?## "}; \
		/^[a-zA-Z_-]+:.*?## / \
		{printf "$(GREEN)%-20s$(RESET) %s\n", $$1, $$2}' \
		$(MAKEFILE_LIST)
	@echo -e "\n$(BLUE)Current configuration:$(RESET)"
	@echo -e "$(GREEN)Environment:$(RESET) $(YELLOW)$(ENV_NAME)$(RESET)"
	@echo -e "$(GREEN)Config source:$(RESET) $(ENV_SOURCE)"
	@echo -e "$(GREEN)Git branch:$(RESET) $(YELLOW)$(CURRENT_BRANCH)$(RESET)"
	@echo -e "$(GREEN)Active server profile:$(RESET) $(YELLOW)$(PROFILE)$(RESET)"

# Отладочные команды
##############################################################################

print-%:  ## Print any make variable
	@echo -e "$(BLUE)$*$(RESET) = $($*)"

info-%:  ## Show commands that would be executed for a target
	@$(MAKE) --dry-run --always-make $* | grep -v "info"

.SILENT: help print-% info-%