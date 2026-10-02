import * as React from "react"
import { Check, ChevronDown, Search, X } from "lucide-react"
import { Popover } from "radix-ui"

import { Input } from "@/components/ui/input"
import { cn } from "@/lib/utils"

export interface MultiSelectOption {
	value: string
	label?: string
	disabled?: boolean
}

interface MultiSelectProps {
	options: MultiSelectOption[]
	value: string[]
	onValueChange: (value: string[]) => void
	name?: string
	placeholder?: string
	searchPlaceholder?: string
	emptyMessage?: string
	disabled?: boolean
	className?: string
}

function MultiSelect({
	options,
	value,
	onValueChange,
	name,
	placeholder = "Select options",
	searchPlaceholder = "Search options...",
	emptyMessage = "No options found.",
	disabled = false,
	className,
}: MultiSelectProps) {
	const [search, setSearch] = React.useState("")
	const selectedOptions = options.filter((option) => value.includes(option.value))
	const filteredOptions = options.filter((option) =>
		(option.label ?? option.value).toLocaleLowerCase().includes(search.toLocaleLowerCase())
	)

	function toggleOption(option: MultiSelectOption) {
		if (option.disabled) {
			return
		}

		onValueChange(
			value.includes(option.value)
				? value.filter((selectedValue) => selectedValue !== option.value)
				: [...value, option.value]
		)
	}

	return (
		<div className={cn("space-y-2", className)}>
			{name && value.map((selectedValue) => (
				<input key={selectedValue} type="hidden" name={name} value={selectedValue} />
			))}
			{selectedOptions.length > 0 && (
				<div className="flex flex-wrap gap-1.5">
					{selectedOptions.map((option) => (
						<span
							key={option.value}
							className="inline-flex h-6 items-center gap-1 rounded-md bg-secondary px-2 text-xs text-secondary-foreground"
						>
							{option.label ?? option.value}
							<button
								type="button"
								aria-label={`Remove ${option.label ?? option.value}`}
								disabled={disabled || option.disabled}
								onClick={() => toggleOption(option)}
								className="rounded-sm text-muted-foreground hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:pointer-events-none"
							>
								<X className="size-3" />
							</button>
						</span>
					))}
				</div>
			)}
			<Popover.Root onOpenChange={(open) => !open && setSearch("")}>
				<Popover.Trigger asChild>
					<button
						type="button"
						disabled={disabled}
						className="flex h-9 w-full items-center justify-between gap-2 rounded-md border border-input bg-transparent px-2.5 py-1 text-left text-sm shadow-xs outline-none transition-[color,box-shadow] focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50"
					>
						<span className={cn("truncate", selectedOptions.length === 0 && "text-muted-foreground")}>
							{selectedOptions.length > 0
								? `${selectedOptions.length} selected`
								: placeholder}
						</span>
						<ChevronDown className="size-4 shrink-0 text-muted-foreground" />
					</button>
				</Popover.Trigger>
				<Popover.Portal>
					<Popover.Content
						align="start"
						sideOffset={4}
						className="z-50 w-[var(--radix-popover-trigger-width)] rounded-md border border-border-card bg-surface-card p-1 shadow-lg"
					>
						<div className="relative p-1">
							<Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
							<Input
								value={search}
								onChange={(event) => setSearch(event.target.value)}
								placeholder={searchPlaceholder}
								className="pl-8"
								autoFocus
							/>
						</div>
						<div role="listbox" aria-multiselectable="true" className="max-h-56 overflow-y-auto p-1">
							{filteredOptions.length === 0 ? (
								<p className="px-2 py-3 text-center text-sm text-muted-foreground">{emptyMessage}</p>
							) : (
								filteredOptions.map((option) => {
									const isSelected = value.includes(option.value)

									return (
										<button
											key={option.value}
											type="button"
											role="option"
											aria-selected={isSelected}
											disabled={option.disabled}
											onClick={() => toggleOption(option)}
											className="flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-left text-sm text-text-secondary outline-none hover:bg-surface-page focus-visible:bg-surface-page disabled:pointer-events-none disabled:opacity-50"
										>
											<span className="flex size-4 items-center justify-center rounded-sm border border-input">
												{isSelected && <Check className="size-3 text-primary" />}
											</span>
											<span className="truncate">{option.label ?? option.value}</span>
										</button>
									)
								})
							)}
						</div>
					</Popover.Content>
				</Popover.Portal>
			</Popover.Root>
		</div>
	)
}

export { MultiSelect }