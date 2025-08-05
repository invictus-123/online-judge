package com.online.judge.backend.service;

import static com.online.judge.backend.factory.ProblemFactory.createProblem;
import static com.online.judge.backend.factory.TagFactory.createTag;
import static com.online.judge.backend.factory.TestCaseFactory.createTestCase;
import static com.online.judge.backend.factory.UiFactory.createProblemDetailsUi;
import static com.online.judge.backend.factory.UiFactory.createTestCaseUi;
import static com.online.judge.backend.factory.UserFactory.createUser;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.verifyNoInteractions;
import static org.mockito.Mockito.when;

import com.online.judge.backend.dto.filter.ProblemFilterRequest;
import com.online.judge.backend.dto.request.CreateProblemRequest;
import com.online.judge.backend.dto.request.CreateTestCaseRequest;
import com.online.judge.backend.dto.ui.ProblemDetailsUi;
import com.online.judge.backend.dto.ui.ProblemSummaryUi;
import com.online.judge.backend.exception.ProblemNotFoundException;
import com.online.judge.backend.exception.UserNotAuthorizedException;
import com.online.judge.backend.model.Problem;
import com.online.judge.backend.model.TestCase;
import com.online.judge.backend.model.User;
import com.online.judge.backend.model.shared.ProblemDifficulty;
import com.online.judge.backend.model.shared.ProblemTag;
import com.online.judge.backend.model.shared.UserRole;
import com.online.judge.backend.repository.ProblemRepository;
import com.online.judge.backend.util.UserUtil;
import java.util.List;
import java.util.Optional;
import java.util.stream.Stream;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;
import org.mockito.ArgumentCaptor;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.PageImpl;
import org.springframework.data.domain.PageRequest;
import org.springframework.data.domain.Pageable;
import org.springframework.data.domain.Sort;
import org.springframework.data.jpa.domain.Specification;

@ExtendWith(MockitoExtension.class)
class ProblemServiceTest {
	private static final int PAGE_SIZE = 20;
	private static final String PROBLEM_TITLE = "New Problem";
	private static final String PROBLEM_STATEMENT = "Statement";
	private static final double TIME_LIMIT = 1.0;
	private static final int MEMORY_LIMIT = 256;
	private static final ProblemDifficulty PROBLEM_DIFFICULTY = ProblemDifficulty.EASY;
	private static final List<ProblemTag> PROBLEM_TAGS = List.of(ProblemTag.ARRAY, ProblemTag.GREEDY);

	@Mock
	private ProblemRepository problemRepository;

	@Mock
	private UserUtil userUtil;

	private ProblemService problemService;

	@BeforeEach
	void setUp() {
		problemService = new ProblemService(problemRepository, userUtil, PAGE_SIZE);
	}

	@Test
	void getProblemDetailsById_whenProblemExists_returnsProblemDetails() {
		Long problemId = 1L;
		Problem mockProblem = createProblem();
		mockProblem.setId(problemId);
		TestCase sampleTestCase = createTestCase(mockProblem, true);
		TestCase hiddenTestCase = createTestCase(mockProblem, false);
		mockProblem.setTestCases(List.of(sampleTestCase, hiddenTestCase));
		when(problemRepository.findById(problemId)).thenReturn(Optional.of(mockProblem));
		ProblemDetailsUi expectedProblemDetails = new ProblemDetailsUi(
				mockProblem.getId(),
				mockProblem.getTitle(),
				mockProblem.getStatement(),
				mockProblem.getTimeLimitSecond(),
				mockProblem.getMemoryLimitMb(),
				mockProblem.getDifficulty(),
				List.of(ProblemTag.ARRAY, ProblemTag.GREEDY),
				List.of(createTestCaseUi(sampleTestCase)));

		ProblemDetailsUi result = problemService.getProblemDetailsById(problemId);

		assertNotNull(result);
		assertEquals(expectedProblemDetails, result);
	}

	@Test
	void getProblemDetailsById_whenProblemDoesNotExist_throwsProblemNotFoundException() {
		Long problemId = 99L;
		when(problemRepository.findById(problemId)).thenReturn(Optional.empty());

		ProblemNotFoundException exception =
				assertThrows(ProblemNotFoundException.class, () -> problemService.getProblemDetailsById(problemId));

		assertEquals("Problem with ID " + problemId + " not found", exception.getMessage());
	}

	@Test
	void createProblem_whenUserIsAdmin_createsAndReturnsProblem() {
		User adminUser = createUser("admin", UserRole.ADMIN);
		when(userUtil.getCurrentAuthenticatedUser()).thenReturn(adminUser);
		CreateProblemRequest request = new CreateProblemRequest(
				PROBLEM_TITLE,
				PROBLEM_STATEMENT,
				PROBLEM_DIFFICULTY,
				TIME_LIMIT,
				MEMORY_LIMIT,
				PROBLEM_TAGS,
				List.of(new CreateTestCaseRequest(
						"Test1 Input", "Test1 Output", /*isSample=*/ true, "Test1 Explanation")));
		Problem problem = createProblemFromRequest(request);
		problem.setCreatedBy(adminUser);
		when(problemRepository.save(any(Problem.class))).thenReturn(problem);
		ProblemDetailsUi expectedProblemDetails = createProblemDetailsUi(problem);

		ProblemDetailsUi result = problemService.createProblem(request);

		ArgumentCaptor<Problem> problemCaptor = ArgumentCaptor.forClass(Problem.class);
		verify(problemRepository).save(problemCaptor.capture());
		Problem capturedProblem = problemCaptor.getValue();
		assertEquals(problem.getTitle(), capturedProblem.getTitle());
		assertEquals(problem.getStatement(), capturedProblem.getStatement());
		assertEquals(problem.getDifficulty(), capturedProblem.getDifficulty());
		assertEquals(problem.getTimeLimitSecond(), capturedProblem.getTimeLimitSecond());
		assertEquals(problem.getMemoryLimitMb(), capturedProblem.getMemoryLimitMb());
		assertEquals(problem.getTags().size(), capturedProblem.getTags().size());
		assertEquals(
				problem.getTestCases().size(), capturedProblem.getTestCases().size());
		assertEquals(expectedProblemDetails, result);
	}

	@Test
	void createProblem_whenUserIsNotAdmin_throwsUserNotAuthorizedException() {
		User regularUser = createUser();
		when(userUtil.getCurrentAuthenticatedUser()).thenReturn(regularUser);

		CreateProblemRequest request = new CreateProblemRequest(
				PROBLEM_TITLE,
				PROBLEM_STATEMENT,
				PROBLEM_DIFFICULTY,
				TIME_LIMIT,
				MEMORY_LIMIT,
				PROBLEM_TAGS,
				List.of(new CreateTestCaseRequest(
						"Input 1", "Output 1", /*isSample=*/ false, /* explanation= */ null)));

		assertThrows(
				UserNotAuthorizedException.class,
				() -> problemService.createProblem(request),
				"User is not authorized to create problems.");
		verifyNoInteractions(problemRepository);
	}

	static Stream<Arguments> listProblemsScenarios() {
		return Stream.of(
				// Basic scenarios - no filters
				Arguments.of("no filters - problems exist", null, null, 1, 2, false),
				Arguments.of("no filters - no problems exist", null, null, 1, 0, true),
				Arguments.of("empty filters", List.of(), List.of(), 1, 2, false),

				// Difficulty filtering
				Arguments.of("single difficulty filter", List.of(ProblemDifficulty.EASY), null, 1, 1, false),
				Arguments.of(
						"multiple difficulties filter",
						List.of(ProblemDifficulty.EASY, ProblemDifficulty.HARD),
						null,
						1,
						2,
						false),
				Arguments.of("difficulty filter - no matches", List.of(ProblemDifficulty.HARD), null, 1, 0, true),

				// Tag filtering
				Arguments.of("single tag filter", null, List.of(ProblemTag.ARRAY), 1, 1, false),
				Arguments.of("multiple tags filter", null, List.of(ProblemTag.TREE, ProblemTag.GRAPH), 1, 2, false),

				// Combined filtering
				Arguments.of(
						"combined difficulty and tag filtering",
						List.of(ProblemDifficulty.EASY),
						List.of(ProblemTag.ARRAY),
						1,
						1,
						false),

				// Pagination
				Arguments.of("custom page number", null, null, 3, 1, false),
				Arguments.of("page 5 with results", null, null, 5, 2, false));
	}

	@ParameterizedTest(name = "listProblems - {0}")
	@MethodSource("listProblemsScenarios")
	void listProblems_allScenarios(
			String scenario,
			List<ProblemDifficulty> difficulties,
			List<ProblemTag> tags,
			int page,
			int expectedResultCount,
			boolean isEmpty) {

		ProblemFilterRequest filterRequest = new ProblemFilterRequest(difficulties, tags, page);

		// Create appropriate mock problems based on scenario
		List<Problem> problems;
		if (isEmpty) {
			problems = List.of();
		} else if (expectedResultCount == 1) {
			if (difficulties != null && difficulties.contains(ProblemDifficulty.EASY)) {
				problems = List.of(createProblem(
						"Problem 1",
						"Statement 1",
						1.0,
						256,
						ProblemDifficulty.EASY,
						createUser(),
						List.of(ProblemTag.ARRAY)));
			} else if (tags != null && tags.contains(ProblemTag.ARRAY)) {
				problems = List.of(createProblem(
						"Problem 1",
						"Statement 1",
						1.0,
						256,
						ProblemDifficulty.EASY,
						createUser(),
						List.of(ProblemTag.ARRAY)));
			} else {
				problems = List.of(createProblem());
			}
		} else if (expectedResultCount == 2) {
			if (difficulties != null
					&& difficulties.contains(ProblemDifficulty.EASY)
					&& difficulties.contains(ProblemDifficulty.HARD)) {
				problems = List.of(
						createProblem(
								"Problem 1",
								"Statement 1",
								1.0,
								256,
								ProblemDifficulty.EASY,
								createUser(),
								List.of(ProblemTag.ARRAY)),
						createProblem(
								"Problem 2",
								"Statement 2",
								1.0,
								256,
								ProblemDifficulty.HARD,
								createUser(),
								List.of(ProblemTag.STRING)));
			} else if (tags != null && tags.contains(ProblemTag.TREE) && tags.contains(ProblemTag.GRAPH)) {
				problems = List.of(
						createProblem(
								"Problem 1",
								"Statement 1",
								1.0,
								256,
								ProblemDifficulty.EASY,
								createUser(),
								List.of(ProblemTag.TREE)),
						createProblem(
								"Problem 2",
								"Statement 2",
								1.0,
								256,
								ProblemDifficulty.MEDIUM,
								createUser(),
								List.of(ProblemTag.GRAPH)));
			} else {
				problems = List.of(createProblem(), createProblem());
			}
		} else {
			problems = List.of(createProblem());
		}

		Pageable expectedPageable =
				PageRequest.of(page - 1, PAGE_SIZE, Sort.by("createdAt").descending());
		Page<Problem> problemPage =
				isEmpty ? Page.empty() : new PageImpl<>(problems, expectedPageable, problems.size());
		when(problemRepository.findAll(any(Specification.class), any(Pageable.class)))
				.thenReturn(problemPage);

		List<ProblemSummaryUi> result = problemService.listProblems(filterRequest);

		// Assertions
		assertNotNull(result);
		assertEquals(expectedResultCount, result.size());
		if (isEmpty) {
			assertTrue(result.isEmpty());
		}
		verify(problemRepository).findAll(any(Specification.class), any(Pageable.class));

		// Verify pagination for custom page scenarios
		if (page == 3) {
			ArgumentCaptor<Pageable> pageableCaptor = ArgumentCaptor.forClass(Pageable.class);
			verify(problemRepository).findAll(any(Specification.class), pageableCaptor.capture());
			Pageable capturedPageable = pageableCaptor.getValue();
			assertEquals(2, capturedPageable.getPageNumber()); // Page 3 should be index 2
			assertEquals(PAGE_SIZE, capturedPageable.getPageSize());
		}
	}

	private Problem createProblemFromRequest(CreateProblemRequest request) {
		Problem problem = new Problem();
		problem.setTitle(request.title());
		problem.setStatement(request.statement());
		problem.setDifficulty(request.difficulty());
		problem.setTimeLimitSecond(request.timeLimit());
		problem.setMemoryLimitMb(request.memoryLimit());
		problem.setTags(
				request.tags().stream().map(tag -> createTag(problem, tag)).toList());
		problem.setTestCases(request.testCases().stream()
				.map(testCaseRequest -> createTestCase(
						problem,
						testCaseRequest.isSample(),
						testCaseRequest.input(),
						testCaseRequest.expectedOutput(),
						testCaseRequest.explanation()))
				.toList());
		return problem;
	}
}
