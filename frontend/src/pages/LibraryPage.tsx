import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useQueryClient } from "@tanstack/react-query";
import { useDocuments, useUploadDocument } from "../hooks/useDocuments";
import UploadDialog from "../components/UploadDialog";
import type { DocumentSummary } from "../api/documents";

function LibraryHeader({
	isUploadOpen,
	onOpen,
}: {
	isUploadOpen: boolean;
	onOpen: () => void;
}) {
	return (
		<div className="flex justify-between items-center mb-4">
			<h1 className="text-2xl font-bold">My Library</h1>
			{!isUploadOpen && (
				<button
					onClick={onOpen}
					className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
				>
					Upload
				</button>
			)}
		</div>
	);
}

export default function LibraryPage() {
	const navigate = useNavigate();
	const queryClient = useQueryClient();
	const { data: documents, isLoading } = useDocuments();
	const uploadMutation = useUploadDocument();
	const [isUploadOpen, setIsUploadOpen] = useState(false);

	const handleUpload = async (file: File) => {
		const result = await uploadMutation.mutateAsync(file);
		queryClient.setQueryData<DocumentSummary[]>(["documents"], (old) => [
			...(old ?? []),
			result,
		]);
		setIsUploadOpen(false);
	};

	if (isLoading) {
		return (
			<div className="max-w-6xl mx-auto p-6">
				<LibraryHeader
					isUploadOpen={isUploadOpen}
					onOpen={() => setIsUploadOpen(true)}
				/>
				<p className="text-gray-500">Loading...</p>
				<UploadDialog
					open={isUploadOpen}
					onClose={() => setIsUploadOpen(false)}
					onUpload={handleUpload}
				/>
			</div>
		);
	}

	if (documents && documents.length === 0) {
		return (
			<div className="max-w-6xl mx-auto p-6">
				<LibraryHeader
					isUploadOpen={isUploadOpen}
					onOpen={() => setIsUploadOpen(true)}
				/>
				<p className="text-gray-500">
					No documents yet. Upload your first EPUB.
				</p>
				<UploadDialog
					open={isUploadOpen}
					onClose={() => setIsUploadOpen(false)}
					onUpload={handleUpload}
				/>
			</div>
		);
	}

	return (
		<div className="max-w-6xl mx-auto p-6">
			<LibraryHeader
				isUploadOpen={isUploadOpen}
				onOpen={() => setIsUploadOpen(true)}
			/>
			<div className="grid gap-4">
				{documents?.map((doc) => (
					<DocumentCard
						key={doc.id}
						document={doc}
						onClick={() => navigate(`/read/${doc.id}/0`)}
					/>
				))}
			</div>
			<UploadDialog
				open={isUploadOpen}
				onClose={() => setIsUploadOpen(false)}
				onUpload={handleUpload}
			/>
		</div>
	);
}

function DocumentCard({
	document,
	onClick,
}: {
	document: DocumentSummary;
	onClick: () => void;
}) {
	return (
		<div
			className="border rounded-lg p-4 hover:shadow-md cursor-pointer"
			onClick={onClick}
		>
			<h2 className="font-semibold text-lg">{document.title}</h2>
			<div className="flex gap-2 mt-1 text-sm text-gray-500">
				<span className="uppercase">{document.file_type}</span>
				<span>{document.chapter_count} chapters</span>
				{document.language && <span>{document.language}</span>}
				<span className={statusColor(document.status)}>{document.status}</span>
			</div>
		</div>
	);
}

function statusColor(status: string): string {
	switch (status) {
		case "ready":
			return "text-green-600";
		case "error":
			return "text-red-600";
		default:
			return "text-yellow-600";
	}
}
