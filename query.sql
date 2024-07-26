

SELECT
    date(series.date),
    member_count.new_members,
    group_days.new_groups,
    metric_days.new_metrics,
    measurement_days.users_with_measurements
FROM
    generate_series(current_date - INTERVAL '7 days', current_date, '1 day'::interval) AS series(date)
LEFT JOIN
    (select count(*) new_members, date(create_date) create_date from members group by 2) as member_count 
ON
    series.date = member_count.create_date
LEFT JOIN
    (select count(*) new_groups, date(create_date) create_date from families group by 2) as group_days 
ON
    series.date = group_days.create_date
LEFT JOIN
    (select count(*) new_metrics, date(most_recent_measurement_date) create_date from member_program_metrics group by 2) as metric_days 
ON
    series.date = metric_days.create_date
LEFT JOIN
    (select count(*) users_with_measurements, create_date from (select  distinct member_id, date(create_date) create_date from member_measurements group by 1,2) as foo  group by 2 order by 2 desc) as measurement_days 
ON
    series.date = measurement_days.create_date



ORDER BY
    series.date desc