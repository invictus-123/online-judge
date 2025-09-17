package com.online.judge.executorv2.model.message;

import jakarta.validation.constraints.NotNull;
import java.util.UUID;

public record TestCaseMessage(@NotNull UUID testCaseId, @NotNull String input, @NotNull String output) {}