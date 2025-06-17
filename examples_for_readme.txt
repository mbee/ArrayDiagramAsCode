# This file contains table syntax examples for the README.md

# Example 1: Simple Table
main_table: [simple-table]
table: [simple-table] My First Table {bg_table:#F5F5F5, bg_cell:#FFFFFF, edge_color:#666666}
[HeaderCol1] First Column Header | [HeaderCol2] Second Column Header
Row 1, Cell 1 | Row 1, Cell 2
Row 2, Cell 1 | Row 2, Cell 2 has\nmultiple lines of text.

# Example 2: Styling and Spanning
main_table: [styling-spanning-table]
table: [styling-spanning-table] Styling and Spanning Demo {bg_table:#E0FFE0, edge_color:#006400, edge_thickness:2}
Feature ::colspan=3:: {bg:#A0D0A0}
[Type] Type {bg:#C0E0C0} | [Description] Description {bg:#C0E0C0} | [Notes] Notes {bg:#C0E0C0}
Rowspan Example ::rowspan=2:: {bg:#D0F0D0} | This cell spans two rows. | Initial note for rowspan.
                                          | Spanned content replaces this cell. | Second note for rowspan.
Colspan Example {bg:#D0F0D0} | This cell uses colspan. ::colspan=2:: {bg:#E0F0E0} |
Individual Cell Style | Normal | Special Cell {bg:#FFDAB9}

# Example 3: Fixed Cell Dimensions
main_table: [fixed-dimensions-table]
table: [fixed-dimensions-table] Fixed Cell Dimensions {bg_table:#FFF5E0, bg_cell:#FFFDF5}
Description | Cell with Fixed Dimensions
Short Text | This cell has a fixed width of 150px and a fixed height of 60px. The text inside will wrap, and if it's too long, it might be clipped. ::fixed_width=150:: ::fixed_height=60::
More Text | Fixed Width only ::fixed_width=100::
Even More | Fixed Height only ::fixed_height=40:: This text might be clipped if it's too long for the height.
Another Row | Both fixed: ::fixed_width=80:: ::fixed_height=30::

# Example 4: Basic Nested Table
main_table: [basic-nested-outer]
table: [basic-nested-outer] Outer Table with Nested Content {bg_table:#E6E6FA}
Section | Details Area ::fixed_width=250:: ::fixed_height=120::
Alpha | Contains details from 'inner-table-1' ::table=inner-table-1::
Beta  | Also contains 'inner-table-1', but with different parent cell text. ::table=inner-table-1:: ::inner_align=center:: ::inner_scale=fit_both::

table: [inner-table-1] Inner Table One {bg_table:#FFFACD, bg_cell:#FFFFE0, edge_color:#BDB76B}
Key | Value
PropA | Value A
PropB | Value B \n (on two lines)

# Example 5: Nested Table - Scaling Options
# To generate individual PNGs for scale modes, we'll define multiple outer tables,
# each referencing the same inner table but with different scaling.

# 5a: inner_scale=none (default)
main_table: [nested-scale-none-outer]
table: [nested-scale-none-outer] Nested: Scale "none" {bg_table:#ADD8E6}
Parent Cell (200x80) ::fixed_width=200:: ::fixed_height=80:: | Description
::table=inner-table-for-scaling:: | Default scaling (none). Inner table might be clipped or smaller than cell.

# 5b: inner_scale=fit_width
main_table: [nested-scale-fitwidth-outer]
table: [nested-scale-fitwidth-outer] Nested: Scale "fit_width" {bg_table:#ADD8E6}
Parent Cell (200x80) ::fixed_width=200:: ::fixed_height=80:: | Description
::table=inner-table-for-scaling:: ::inner_scale=fit_width:: | Scales to fit width. Height adjusts by aspect ratio.

# 5c: inner_scale=fit_height
main_table: [nested-scale-fitheight-outer]
table: [nested-scale-fitheight-outer] Nested: Scale "fit_height" {bg_table:#ADD8E6}
Parent Cell (200x80) ::fixed_width=200:: ::fixed_height=80:: | Description
::table=inner-table-for-scaling:: ::inner_scale=fit_height:: | Scales to fit height. Width adjusts by aspect ratio.

# 5d: inner_scale=fit_both
main_table: [nested-scale-fitboth-outer]
table: [nested-scale-fitboth-outer] Nested: Scale "fit_both" {bg_table:#ADD8E6}
Parent Cell (200x80) ::fixed_width=200:: ::fixed_height=80:: | Description
::table=inner-table-for-scaling:: ::inner_scale=fit_both:: | Scales to fit both width and height, maintaining aspect ratio.

# 5e: inner_scale=fill_stretch
main_table: [nested-scale-fillstretch-outer]
table: [nested-scale-fillstretch-outer] Nested: Scale "fill_stretch" {bg_table:#ADD8E6}
Parent Cell (200x80) ::fixed_width=200:: ::fixed_height=80:: | Description
::table=inner-table-for-scaling:: ::inner_scale=fill_stretch:: | Stretches to fill entire cell, ignoring aspect ratio.

table: [inner-table-for-scaling] Inner Table (for scaling demos) {bg_table:#FFFFE0, edge_color:#FFD700}
Column X | Column Y
Data 123 | Data 456
More Data | And More

# Example 6: Nested Table - Alignment Options
# Similar to scaling, we'll use multiple outer tables for alignment.
# Using a slightly smaller fixed size for parent cell to make alignment more visible.

# 6a: inner_align=top_left (default)
main_table: [nested-align-tl-outer]
table: [nested-align-tl-outer] Nested: Align "top_left" {bg_table:#E0FFFF}
Parent Cell (220x100) ::fixed_width=220:: ::fixed_height=100:: | Description
::table=inner-table-for-aligning:: ::inner_scale=none:: | Default align (top_left). Using 'none' scale to show natural size.

# 6b: inner_align=center
main_table: [nested-align-center-outer]
table: [nested-align-center-outer] Nested: Align "center" {bg_table:#E0FFFF}
Parent Cell (220x100) ::fixed_width=220:: ::fixed_height=100:: | Description
::table=inner-table-for-aligning:: ::inner_scale=none:: ::inner_align=center:: | Centered alignment.

# 6c: inner_align=bottom_right
main_table: [nested-align-br-outer]
table: [nested-align-br-outer] Nested: Align "bottom_right" {bg_table:#E0FFFF}
Parent Cell (220x100) ::fixed_width=220:: ::fixed_height=100:: | Description
::table=inner-table-for-aligning:: ::inner_scale=none:: ::inner_align=bottom_right:: | Bottom-right alignment.

table: [inner-table-for-aligning] Inner Table (for aligning demos) {bg_table:#FAFAD2, edge_color:#B0E0E6}
Info A | Info B
X | Y
