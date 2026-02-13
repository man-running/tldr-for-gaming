import NotFound from "@/app/not-found";

export default function PaperNotFound() {
	return (
		<NotFound
			title="404"
			message="Paper Not Found"
			context="The requested research paper could not be found or is not available. The paper might not be indexed yet or the arXiv ID might be incorrect."
			showBackButton={true}
		/>
	);
}
