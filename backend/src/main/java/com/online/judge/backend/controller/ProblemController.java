package com.online.judge.backend.controller;

import com.online.judge.backend.annotation.RateLimit;
import com.online.judge.backend.dto.filter.ProblemFilterRequest;
import com.online.judge.backend.dto.request.CreateProblemRequest;
import com.online.judge.backend.dto.response.CreateProblemResponse;
import com.online.judge.backend.dto.response.GetProblemByIdResponse;
import com.online.judge.backend.dto.response.ListProblemsResponse;
import com.online.judge.backend.model.shared.ProblemDifficulty;
import com.online.judge.backend.model.shared.ProblemTag;
import com.online.judge.backend.service.ProblemService;
import jakarta.validation.Valid;
import java.util.List;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

/** REST controller for handling problem-related requests. */
@RestController
@RequestMapping("/api/v1/problems")
public class ProblemController {
	private static final Logger logger = LoggerFactory.getLogger(ProblemController.class);

	private final ProblemService problemService;

	public ProblemController(ProblemService problemService) {
		this.problemService = problemService;
	}

	/**
	 * Handles GET requests to fetch a list of all problems with pagination and optional filtering,
	 * sorted by creation date in descending order.
	 *
	 * @param page The page number to retrieve (default is 1).
	 * @param difficulties List of difficulties to filter by (optional).
	 * @param tags List of tags to filter by (optional).
	 * @return A ResponseEntity containing a paginated list of problems matching the filters.
	 */
	@GetMapping("/list")
	public ResponseEntity<ListProblemsResponse> listProblems(
			@RequestParam(defaultValue = "1") int page,
			@RequestParam(required = false, name = "difficulty") List<ProblemDifficulty> difficulties,
			@RequestParam(required = false, name = "tag") List<ProblemTag> tags) {
		ProblemFilterRequest filterRequest = new ProblemFilterRequest(difficulties, tags, page);
		logger.info("Received call to fetch problems with filters={}", filterRequest);

		ListProblemsResponse response = new ListProblemsResponse(problemService.listProblems(filterRequest));
		return ResponseEntity.ok(response);
	}

	/**
	 * Handles GET requests to fetch the details of a specific problem by its ID.
	 *
	 * @param id
	 *            The ID of the problem to retrieve.
	 * @return A ResponseEntity containing the ProblemDetailsUi DTO if found, or a
	 *         404 Not Found status if the problem does not exist.
	 */
	@GetMapping("/{id}")
	public ResponseEntity<GetProblemByIdResponse> getProblemById(@PathVariable Long id) {
		logger.info("Received call to fetch problem with ID {}", id);

		GetProblemByIdResponse response = new GetProblemByIdResponse(problemService.getProblemDetailsById(id));
		return ResponseEntity.ok(response);
	}

	/**
	 * Handles POST requests to create a new problem.
	 * Only accessible by users with the 'ADMIN' role.
	 *
	 * @param request The request body containing the problem details.
	 * @return A ResponseEntity with the created problem and a 201 Created status. Throws a
	 *         401 error if the user is not authorized to create a problem.
	 */
	@PostMapping
	@RateLimit(apiType = "create-problem", capacity = 10, refillPeriodMinutes = 1)
	public ResponseEntity<CreateProblemResponse> createProblem(@Valid @RequestBody CreateProblemRequest request) {
		logger.info("Received request to create a new problem: {}", request);

		CreateProblemResponse response = new CreateProblemResponse(problemService.createProblem(request));
		return new ResponseEntity<>(response, HttpStatus.CREATED);
	}
}
