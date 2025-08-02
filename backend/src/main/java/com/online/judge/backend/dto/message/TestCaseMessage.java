package com.online.judge.backend.dto.message;

import jakarta.validation.constraints.NotNull;
import java.util.UUID;

/**
 * DTO representing the data for a particular test case.
 */
public record TestCaseMessage(@NotNull UUID testCaseId, @NotNull String input, @NotNull String output) {}
