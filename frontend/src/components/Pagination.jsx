import React from 'react';
import { ChevronLeft, ChevronRight } from 'lucide-react';

const Pagination = ({ currentPage, totalPages, onPageChange }) => {
    if (totalPages <= 1) return null;

    return (
        <div className="flex justify-between items-center bg-white border-t-2 border-ink-800/5 p-3">
            <div className="text-[10px] uppercase tracking-widest text-ink-800/40 font-bold">
                Showing Page {currentPage} of {totalPages}
            </div>
            <div className="flex gap-2">
                <button
                    onClick={() => onPageChange(currentPage - 1)}
                    disabled={currentPage === 1}
                    className="flex items-center gap-1 px-3 py-1.5 rounded bg-parchment-100 text-ink-800 text-[10px] font-bold uppercase tracking-wide hover:bg-ink-800 hover:text-white transition-all disabled:opacity-30 disabled:hover:bg-parchment-100 disabled:hover:text-ink-800 border border-ink-800/10 shadow-sm"
                >
                    <ChevronLeft size={12} /> Previous
                </button>
                <div className="flex gap-1">
                    {[...Array(totalPages)].map((_, i) => (
                        <button
                            key={i + 1}
                            onClick={() => onPageChange(i + 1)}
                            className={`w-7 h-7 flex items-center justify-center rounded text-[10px] font-bold transition-all ${currentPage === i + 1
                                    ? 'bg-ink-800 text-white shadow-sm'
                                    : 'bg-transparent text-ink-800/60 hover:bg-parchment-100'
                                }`}
                        >
                            {i + 1}
                        </button>
                    ))}
                </div>
                <button
                    onClick={() => onPageChange(currentPage + 1)}
                    disabled={currentPage === totalPages}
                    className="flex items-center gap-1 px-3 py-1.5 rounded bg-parchment-100 text-ink-800 text-[10px] font-bold uppercase tracking-wide hover:bg-ink-800 hover:text-white transition-all disabled:opacity-30 disabled:hover:bg-parchment-100 disabled:hover:text-ink-800 border border-ink-800/10 shadow-sm"
                >
                    Next <ChevronRight size={12} />
                </button>
            </div>
        </div>
    );
};

export default Pagination;
