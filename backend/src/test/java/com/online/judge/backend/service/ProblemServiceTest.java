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
import static org.mockito.ArgumentMatchers.eq;
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
import com.online.judge.backend.model.shared.SolvedStatus;
import com.online.judge.backend.model.shared.UserRole;
import com.online.judge.backend.repository.ProblemRepository;
import com.online.judge.backend.util.UserUtil;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.stream.Stream;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;
import org.mockito.ArgumentCaptor;
import org.mockito.ArgumentMatchers;
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
	private static final Problem EASY_ARRAY_PROBLEM = createProblem(ProblemDifficulty.EASY, List.of(ProblemTag.ARRAY));
	private static final Problem MEDIUM_ARRAY_PROBLEM =
			createProblem(ProblemDifficulty.MEDIUM, List.of(ProblemTag.ARRAY));
	private static final Problem MEDIUM_GREEDY_TREE_PROBLEM =
			createProblem(ProblemDifficulty.HARD, List.of(ProblemTag.GREEDY, ProblemTag.TREE));
	private static final Problem MEDIUM_TREE_PROBLEM =
			createProblem(ProblemDifficulty.MEDIUM, List.of(ProblemTag.TREE));

	@Mock
	private ProblemRepository problemRepository;

	@Mock
	private UserUtil userUtil;

	@Mock
	private SolvedStatusService solvedStatusService;

	private ProblemService problemService;

	@BeforeEach
	void setUp() {
		problemService = new ProblemService(problemRepository, solvedStatusService, userUtil, PAGE_SIZE);
	}

	static Stream<Arguments> listProblemsScenarios() {
		return Stream.of(
				// Basic scenarios - no filters
				Arguments.of(
						"no filters - problems exist",
						/* page= */ 1,
						/* difficultes= */ null,
						/* tags= */ null,
						/* problems= */ List.of(
								EASY_ARRAY_PROBLEM,
								MEDIUM_ARRAY_PROBLEM,
								MEDIUM_GREEDY_TREE_PROBLEM,
								MEDIUM_TREE_PROBLEM),
						/* expectedResultCount= */ 4),
				Arguments.of(
						"no filters - no problems exist",
						/* page= */ 1,
						/* difficultes= */ null,
						/* tags= */ null,
						/* problems= */ List.of(),
						/* expectedResultCount= */ 0),
				Arguments.of(
						"empty filters",
						/* page= */ 1,
						/* difficultes= */ List.of(),
						/* tags= */ List.of(),
						/* problems= */ List.of(
								EASY_ARRAY_PROBLEM,
								MEDIUM_ARRAY_PROBLEM,
								MEDIUM_GREEDY_TREE_PROBLEM,
								MEDIUM_TREE_PROBLEM),
						/* expectedResultCount= */ 4),

				// Difficulty filtering
				Arguments.of(
						"single difficulty filter",
						/* page= */ 1,
						/* difficultes= */ List.of(ProblemDifficulty.MEDIUM),
						/* tags= */ List.of(),
						/* problems= */ List.of(MEDIUM_ARRAY_PROBLEM, MEDIUM_TREE_PROBLEM, MEDIUM_GREEDY_TREE_PROBLEM),
						/* expectedResultCount= */ 3),
				Arguments.of(
						"multiple difficulties filter",
						/* page= */ 1,
						/* difficultes= */ List.of(ProblemDifficulty.EASY, ProblemDifficulty.MEDIUM),
						/* tags= */ List.of(),
						/* problems= */ List.of(
								EASY_ARRAY_PROBLEM,
								MEDIUM_ARRAY_PROBLEM,
								MEDIUM_TREE_PROBLEM,
								MEDIUM_GREEDY_TREE_PROBLEM),
						/* expectedResultCount= */ 4),
				Arguments.of(
						"difficulty filter - no matches",
						/* page= */ 1,
						/* difficultes= */ List.of(ProblemDifficulty.HARD),
						/* tags= */ List.of(),
						/* problems= */ List.of(),
						/* expectedResultCount= */ 0),

				// Tag filtering
				Arguments.of(
						"single tag filter",
						/* page= */ 1,
						/* difficultes= */ List.of(),
						/* tags= */ List.of(ProblemTag.ARRAY),
						/* problems= */ List.of(EASY_ARRAY_PROBLEM, MEDIUM_ARRAY_PROBLEM),
						/* expectedResultCount= */ 2),
				Arguments.of(
						"multiple tags filter",
						/* page= */ 1,
						/* difficultes= */ List.of(),
						/* tags= */ List.of(ProblemTag.ARRAY, ProblemTag.GREEDY),
						/* problems= */ List.of(EASY_ARRAY_PROBLEM, MEDIUM_ARRAY_PROBLEM, MEDIUM_GREEDY_TREE_PROBLEM),
						/* expectedResultCount= */ 3),

				// Combined filtering
				Arguments.of(
						"combined difficulty and tag filtering",
						/* page= */ 1,
						/* difficultes= */ List.of(ProblemDifficulty.MEDIUM),
						/* tags= */ List.of(ProblemTag.TREE),
						/* problems= */ List.of(MEDIUM_GREEDY_TREE_PROBLEM, MEDIUM_TREE_PROBLEM),
						/* expectedResultCount= */ 2));
	}

	@ParameterizedTest(name = "listProblems - {0}")
	@MethodSource("listProblemsScenarios")
	void listProblems_allScenarios(
			String scenario,
			int page,
			List<ProblemDifficulty> difficulties,
			List<ProblemTag> tags,
			List<Problem> problems,
			int expectedResultCount) {
		ProblemFilterRequest filterRequest = new ProblemFilterRequest(difficulties, tags, page);
		Pageable expectedPageable =
				PageRequest.of(page - 1, PAGE_SIZE, Sort.by("createdAt").descending());
		Page<Problem> problemPage = new PageImpl<>(problems, expectedPageable, problems.size());
		when(problemRepository.findAll(ArgumentMatchers.<Specification<Problem>>any(), eq(expectedPageable)))
				.thenReturn(problemPage);
		when(userUtil.getCurrentAuthenticatedUserOptional()).thenReturn(Optional.empty());
		List<Long> problemIds = problems.stream().map(Problem::getId).toList();
		Map<Long, SolvedStatus> solvedStatusMap = new HashMap<>();
		problemIds.forEach(id -> solvedStatusMap.put(id, SolvedStatus.UNATTEMPTED));
		when(solvedStatusService.getSolvedStatusForProblems(null, problemIds)).thenReturn(solvedStatusMap);

		List<ProblemSummaryUi> result = problemService.listProblems(filterRequest);

		assertNotNull(result);
		assertEquals(expectedResultCount, result.size());
		result.forEach(problem -> assertEquals(SolvedStatus.UNATTEMPTED, problem.solvedStatus()));
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
		when(userUtil.getCurrentAuthenticatedUserOptional()).thenReturn(Optional.empty());
		when(solvedStatusService.getSolvedStatus(null, problemId)).thenReturn(SolvedStatus.UNATTEMPTED);

		ProblemDetailsUi expectedProblemDetails = new ProblemDetailsUi(
				mockProblem.getId(),
				mockProblem.getTitle(),
				mockProblem.getStatement(),
				mockProblem.getTimeLimitSecond(),
				mockProblem.getMemoryLimitMb(),
				mockProblem.getDifficulty(),
				List.of(ProblemTag.ARRAY, ProblemTag.GREEDY),
				List.of(createTestCaseUi(sampleTestCase)),
				SolvedStatus.UNATTEMPTED);

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
		when(problemRepository.save(ArgumentMatchers.<Problem>any())).thenReturn(problem);
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
