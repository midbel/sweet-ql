with persons as (select first, last from employees) select * from persons;

with persons as (
	select id, first, last from employees
), emails as (
	select id, mail from emails
)
select * from persons t1 join emails t2 on t1.id=t2.id;